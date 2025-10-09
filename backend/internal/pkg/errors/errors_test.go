package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error without wrapped error",
			appError: &AppError{
				Code:    "test_error",
				Message: "This is a test error",
			},
			expected: "test_error: This is a test error",
		},
		{
			name: "error with wrapped error",
			appError: &AppError{
				Code:    "test_error",
				Message: "This is a test error",
				Err:     errors.New("underlying error"),
			},
			expected: "test_error: This is a test error (wrapped: underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	appErr := &AppError{
		Code:    "test_error",
		Message: "This is a test error",
		Err:     underlyingErr,
	}

	unwrapped := appErr.Unwrap()
	assert.Equal(t, underlyingErr, unwrapped)
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewBadRequestError("validation failed")

	result := appErr.WithDetails("field", "email").WithDetails("reason", "invalid format")

	assert.Len(t, result.Details, 2)
	assert.Equal(t, "email", result.Details["field"])
	assert.Equal(t, "invalid format", result.Details["reason"])
}

func TestAppError_WithError(t *testing.T) {
	underlyingErr := errors.New("database connection failed")
	appErr := NewInternalServerError("failed to fetch user")

	result := appErr.WithError(underlyingErr)

	assert.Equal(t, underlyingErr, result.Err)
	assert.Contains(t, result.Error(), "database connection failed")
}

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("invalid input")

	assert.Equal(t, "invalid_request", err.Code)
	assert.Equal(t, "invalid input", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Nil(t, err.Details)
}

func TestNewValidationError(t *testing.T) {
	details := map[string]interface{}{
		"email": "invalid email format",
		"age":   "must be at least 18",
	}
	err := NewValidationError("validation failed", details)

	assert.Equal(t, "validation_error", err.Code)
	assert.Equal(t, "validation failed", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, details, err.Details)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("token expired")

	assert.Equal(t, "unauthorized", err.Code)
	assert.Equal(t, "token expired", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("insufficient permissions")

	assert.Equal(t, "forbidden", err.Code)
	assert.Equal(t, "insufficient permissions", err.Message)
	assert.Equal(t, http.StatusForbidden, err.HTTPStatus)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("User")

	assert.Equal(t, "not_found", err.Code)
	assert.Equal(t, "User not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("email already exists")

	assert.Equal(t, "conflict", err.Code)
	assert.Equal(t, "email already exists", err.Message)
	assert.Equal(t, http.StatusConflict, err.HTTPStatus)
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("too many requests")

	assert.Equal(t, "rate_limit_exceeded", err.Code)
	assert.Equal(t, "too many requests", err.Message)
	assert.Equal(t, http.StatusTooManyRequests, err.HTTPStatus)
}

func TestNewInternalServerError(t *testing.T) {
	err := NewInternalServerError("unexpected error")

	assert.Equal(t, "internal_server_error", err.Code)
	assert.Equal(t, "unexpected error", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("database unavailable")

	assert.Equal(t, "service_unavailable", err.Code)
	assert.Equal(t, "database unavailable", err.Message)
	assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
}

func TestHandleError_WithAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "bad request error",
			err:            NewBadRequestError("invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_request",
			expectedMsg:    "invalid input",
		},
		{
			name:           "unauthorized error",
			err:            NewUnauthorizedError("token expired"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "unauthorized",
			expectedMsg:    "token expired",
		},
		{
			name:           "not found error",
			err:            NewNotFoundError("User"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "not_found",
			expectedMsg:    "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			HandleError(c, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedCode)
			assert.Contains(t, w.Body.String(), tt.expectedMsg)
		})
	}
}

func TestHandleError_WithGenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	genericErr := errors.New("some generic error")
	HandleError(c, genericErr)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_server_error")
	assert.Contains(t, w.Body.String(), "An unexpected error occurred")
	// Should not expose internal error details
	assert.NotContains(t, w.Body.String(), "some generic error")
}

func TestHandleError_WithNilError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, nil)

	// Should not write anything when error is nil
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestHandleError_WithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := map[string]interface{}{
		"field": "email",
		"error": "invalid format",
	}
	err := NewValidationError("validation failed", details)

	HandleError(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")
	assert.Contains(t, w.Body.String(), "validation failed")
	assert.Contains(t, w.Body.String(), "email")
	assert.Contains(t, w.Body.String(), "invalid format")
}

func TestAbortWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handlerExecuted := false
	router.GET("/test", func(c *gin.Context) {
		AbortWithError(c, NewUnauthorizedError("not authenticated"))
		// Check if context was aborted
		if !c.IsAborted() {
			t.Error("Context should be aborted after AbortWithError")
		}
		// This line executes but shouldn't write to response due to abort
		handlerExecuted = true
		if !c.IsAborted() {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.True(t, handlerExecuted, "Handler should have executed")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestErrorResponse(t *testing.T) {
	t.Run("error without details", func(t *testing.T) {
		err := NewBadRequestError("invalid input")
		response := ErrorResponse(err)

		assert.Equal(t, "invalid_request", response["error"])
		assert.Equal(t, "invalid input", response["message"])
		assert.NotContains(t, response, "details")
	})

	t.Run("error with details", func(t *testing.T) {
		details := map[string]interface{}{"field": "email"}
		err := NewValidationError("validation failed", details)
		response := ErrorResponse(err)

		assert.Equal(t, "validation_error", response["error"])
		assert.Equal(t, "validation failed", response["message"])
		assert.Equal(t, details, response["details"])
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap generic error", func(t *testing.T) {
		originalErr := errors.New("database connection failed")
		wrappedErr := WrapError(originalErr, "failed to fetch user", http.StatusInternalServerError, "database_error")

		require.NotNil(t, wrappedErr)
		assert.Equal(t, "database_error", wrappedErr.Code)
		assert.Equal(t, "failed to fetch user", wrappedErr.Message)
		assert.Equal(t, http.StatusInternalServerError, wrappedErr.HTTPStatus)
		assert.Equal(t, originalErr, wrappedErr.Err)
		assert.Contains(t, wrappedErr.Error(), "database connection failed")
	})

	t.Run("wrap nil error", func(t *testing.T) {
		wrappedErr := WrapError(nil, "some message", http.StatusBadRequest, "test_error")
		assert.Nil(t, wrappedErr)
	})

	t.Run("wrap existing AppError", func(t *testing.T) {
		originalErr := NewNotFoundError("User")
		wrappedErr := WrapError(originalErr, "failed to fetch user", http.StatusInternalServerError, "database_error")

		require.NotNil(t, wrappedErr)
		// Should return the original AppError as-is
		assert.Equal(t, originalErr, wrappedErr)
	})
}

func TestIsNotFound(t *testing.T) {
	t.Run("is not found error", func(t *testing.T) {
		err := NewNotFoundError("User")
		assert.True(t, IsNotFound(err))
	})

	t.Run("is not a not found error", func(t *testing.T) {
		err := NewBadRequestError("invalid input")
		assert.False(t, IsNotFound(err))
	})

	t.Run("generic error", func(t *testing.T) {
		err := errors.New("some error")
		assert.False(t, IsNotFound(err))
	})
}

func TestIsUnauthorized(t *testing.T) {
	t.Run("is unauthorized error", func(t *testing.T) {
		err := NewUnauthorizedError("token expired")
		assert.True(t, IsUnauthorized(err))
	})

	t.Run("is not an unauthorized error", func(t *testing.T) {
		err := NewForbiddenError("insufficient permissions")
		assert.False(t, IsUnauthorized(err))
	})
}

func TestIsForbidden(t *testing.T) {
	t.Run("is forbidden error", func(t *testing.T) {
		err := NewForbiddenError("insufficient permissions")
		assert.True(t, IsForbidden(err))
	})

	t.Run("is not a forbidden error", func(t *testing.T) {
		err := NewUnauthorizedError("token expired")
		assert.False(t, IsForbidden(err))
	})
}

func TestIsBadRequest(t *testing.T) {
	t.Run("is bad request error", func(t *testing.T) {
		err := NewBadRequestError("invalid input")
		assert.True(t, IsBadRequest(err))
	})

	t.Run("validation error is also bad request", func(t *testing.T) {
		err := NewValidationError("validation failed", nil)
		assert.True(t, IsBadRequest(err))
	})

	t.Run("is not a bad request error", func(t *testing.T) {
		err := NewNotFoundError("User")
		assert.False(t, IsBadRequest(err))
	})
}

func TestErrorChaining(t *testing.T) {
	// Test that error wrapping works with errors.Is and errors.As
	originalErr := errors.New("original error")
	appErr := NewInternalServerError("wrapped error").WithError(originalErr)

	// Test errors.As
	var extractedAppErr *AppError
	assert.True(t, errors.As(appErr, &extractedAppErr))
	assert.Equal(t, "internal_server_error", extractedAppErr.Code)

	// Test Unwrap chain
	assert.Equal(t, originalErr, errors.Unwrap(appErr))
}

func TestConcurrentDetailsMutation(t *testing.T) {
	// Test that adding details concurrently doesn't cause race conditions
	err := NewBadRequestError("test error")

	done := make(chan bool, 2)

	go func() {
		err.WithDetails("key1", "value1")
		done <- true
	}()

	go func() {
		err.WithDetails("key2", "value2")
		done <- true
	}()

	<-done
	<-done

	// Both details should be present
	assert.Contains(t, err.Details, "key1")
	assert.Contains(t, err.Details, "key2")
}
