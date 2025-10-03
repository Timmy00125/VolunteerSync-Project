package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "JSON format with info level",
			config: Config{
				Level:  "info",
				Format: "json",
			},
		},
		{
			name: "Console format with debug level",
			config: Config{
				Level:  "debug",
				Format: "console",
			},
		},
		{
			name: "With caller information",
			config: Config{
				Level:      "info",
				Format:     "json",
				WithCaller: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.config.Output = buf
			Init(tt.config)

			logger := Get()
			assert.NotNil(t, logger)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"debug level", "debug", "debug"},
		{"info level", "info", "info"},
		{"warn level", "warn", "warn"},
		{"warning level", "warning", "warn"},
		{"error level", "error", "error"},
		{"fatal level", "fatal", "fatal"},
		{"default level", "invalid", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := parseLevel(tt.level)
			assert.Equal(t, tt.expected, level.String())
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, UserIDKey, "user-456")
	ctx = context.WithValue(ctx, OrgIDKey, "org-789")

	logger := Get().WithContext(ctx)
	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "req-123")
	assert.Contains(t, output, "user-456")
	assert.Contains(t, output, "org-789")
	assert.Contains(t, output, "test message")
}

func TestLogger_WithField(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger := Get().WithField("operation", "test_operation")
	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test_operation")
	assert.Contains(t, output, "test message")
}

func TestLogger_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	fields := map[string]interface{}{
		"operation": "test_op",
		"count":     42,
		"active":    true,
	}

	logger := Get().WithFields(fields)
	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test_op")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "true")
}

func TestLogger_WithFields_FiltersPII(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	// Fields that should be filtered out
	fields := map[string]interface{}{
		"password":   "secret123",
		"email":      "user@example.com",
		"phone":      "123-456-7890",
		"safe_field": "this should appear",
	}

	logger := Get().WithFields(fields)
	logger.Info("test message")

	output := buf.String()

	// PII should NOT appear
	assert.NotContains(t, output, "secret123")
	assert.NotContains(t, output, "user@example.com")
	assert.NotContains(t, output, "123-456-7890")

	// Safe field should appear
	assert.Contains(t, output, "this should appear")
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(*Logger)
		level     string
		expectLog bool
	}{
		{
			name: "debug logs at debug level",
			logFunc: func(l *Logger) {
				l.Debug("debug message")
			},
			level:     "debug",
			expectLog: true,
		},
		{
			name: "debug does not log at info level",
			logFunc: func(l *Logger) {
				l.Debug("debug message")
			},
			level:     "info",
			expectLog: false,
		},
		{
			name: "info logs at info level",
			logFunc: func(l *Logger) {
				l.Info("info message")
			},
			level:     "info",
			expectLog: true,
		},
		{
			name: "error logs at info level",
			logFunc: func(l *Logger) {
				l.Error("error message")
			},
			level:     "info",
			expectLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			Init(Config{
				Level:  tt.level,
				Format: "json",
				Output: buf,
			})

			logger := Get()
			tt.logFunc(logger)

			output := buf.String()
			if tt.expectLog {
				assert.NotEmpty(t, output)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_Debugf(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "debug",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.Debugf("test %s %d", "message", 42)

	output := buf.String()
	assert.Contains(t, output, "test message 42")
}

func TestLogger_Infof(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.Infof("test %s %d", "message", 42)

	output := buf.String()
	assert.Contains(t, output, "test message 42")
}

func TestLogger_Warnf(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "warn",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.Warnf("test %s %d", "warning", 42)

	output := buf.String()
	assert.Contains(t, output, "test warning 42")
}

func TestLogger_Errorf(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "error",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.Errorf("test %s %d", "error", 42)

	output := buf.String()
	assert.Contains(t, output, "test error 42")
}

func TestLogger_ErrorWithErr(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "error",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.ErrorWithErr("operation failed", assert.AnError)

	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "error")
}

func TestLogger_LogRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.LogRequest("GET", "/api/v1/users", 200, 150*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "/api/v1/users")
	assert.Contains(t, output, "200")
	assert.Contains(t, output, "HTTP request")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/api/v1/users", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
}

func TestLogger_LogDatabaseQuery(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "debug",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.LogDatabaseQuery("SELECT", "users", 50*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "SELECT")
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "Database query")
}

func TestLogger_LogCacheOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "debug",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.LogCacheOperation("GET", "user:123", true)

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "user:123")
	assert.Contains(t, output, "Cache operation")

	// Parse JSON to verify hit field
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, true, logEntry["hit"])
}

func TestLogger_LogAuthentication(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		level   string
	}{
		{
			name:    "successful authentication",
			success: true,
			level:   "info",
		},
		{
			name:    "failed authentication",
			success: false,
			level:   "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			Init(Config{
				Level:  "debug",
				Format: "json",
				Output: buf,
			})

			logger := Get()
			logger.LogAuthentication("user-123", "login", tt.success)

			output := buf.String()
			assert.Contains(t, output, "user-123")
			assert.Contains(t, output, "login")
			assert.Contains(t, output, "Authentication event")
			assert.Contains(t, output, tt.level)
		})
	}
}

func TestLogger_LogAuthorization(t *testing.T) {
	tests := []struct {
		name    string
		allowed bool
		level   string
	}{
		{
			name:    "allowed authorization",
			allowed: true,
			level:   "info",
		},
		{
			name:    "denied authorization",
			allowed: false,
			level:   "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			Init(Config{
				Level:  "debug",
				Format: "json",
				Output: buf,
			})

			logger := Get()
			logger.LogAuthorization("user-123", "organizations", "create", tt.allowed)

			output := buf.String()
			assert.Contains(t, output, "user-123")
			assert.Contains(t, output, "organizations")
			assert.Contains(t, output, "create")
			assert.Contains(t, output, "Authorization event")
			assert.Contains(t, output, tt.level)
		})
	}
}

func TestIsPIIField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"password field", "password", true},
		{"email field", "email", true},
		{"phone field", "phone", true},
		{"ssn field", "ssn", true},
		{"name field", "name", true},
		{"first_name field", "first_name", true},
		{"address field", "address", true},
		{"ip_address field", "ip_address", true},
		{"safe field", "user_id", false},
		{"safe field", "status", false},
		{"safe field", "count", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPIIField(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUserID(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string ID", "user-123", "user-123"},
		{"int ID", 123, "user_id"},
		{"int64 ID", int64(123), "user_id"},
		{"other type", struct{}{}, "user_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeUserID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeOrgID(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string ID", "org-123", "org-123"},
		{"int ID", 123, "org_id"},
		{"int64 ID", int64(123), "org_id"},
		{"other type", struct{}{}, "org_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeOrgID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGet_ReturnsDefaultWhenNotInitialized(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	logger := Get()
	assert.NotNil(t, logger)

	// Should be able to log without panic (uses stdout by default)
	logger.Info("test message")
}

func TestLogger_NoPIIInLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	// Simulate a scenario where we might accidentally log PII
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")

	logger := Get().WithContext(ctx)

	// This should work without logging the password
	fields := map[string]interface{}{
		"operation": "user_login",
		"password":  "should_not_appear",
		"email":     "should_not_appear@example.com",
	}

	logger = logger.WithFields(fields)
	logger.Info("User login attempt")

	output := buf.String()

	// Verify PII is not in logs
	assert.NotContains(t, output, "should_not_appear")
	assert.NotContains(t, output, "should_not_appear@example.com")

	// Verify safe information is logged
	assert.Contains(t, output, "user_login")
	assert.Contains(t, output, "User login attempt")
}

func TestLogger_StructuredOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger := Get()
	logger.Info("test message")

	// Verify output is valid JSON
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	err := json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, logEntry, "level")
	assert.Contains(t, logEntry, "message")
	assert.Contains(t, logEntry, "time")
	assert.Equal(t, "test message", logEntry["message"])
	assert.Equal(t, "info", logEntry["level"])
}
