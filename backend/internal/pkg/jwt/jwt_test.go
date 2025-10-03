package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotEmpty(t, config.AccessSecret)
	assert.NotEmpty(t, config.RefreshSecret)
	assert.Equal(t, 15*time.Minute, config.AccessTokenExpiry)
	assert.Equal(t, 7*24*time.Hour, config.RefreshTokenExpiry)
	assert.Equal(t, "volunteersync", config.Issuer)
}

func TestNewManager(t *testing.T) {
	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			AccessSecret:       "test-access-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  10 * time.Minute,
			RefreshTokenExpiry: 5 * 24 * time.Hour,
			Issuer:             "test-issuer",
		}

		manager := NewManager(config)
		assert.NotNil(t, manager)
		assert.Equal(t, config, manager.config)
	})

	t.Run("with nil config uses default", func(t *testing.T) {
		manager := NewManager(nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.Equal(t, "volunteersync", manager.config.Issuer)
	})
}

func TestGenerateAccessToken(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("successful generation", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed
		claims, err := manager.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "volunteer", claims.Role)
		assert.Equal(t, AccessTokenType, claims.TokenType)
		assert.Equal(t, "test", claims.Issuer)
		assert.NotEmpty(t, claims.ID)
	})

	t.Run("missing user ID", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("", "volunteer")
		assert.ErrorIs(t, err, ErrMissingUserID)
		assert.Empty(t, token)
	})

	t.Run("missing role", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "")
		assert.ErrorIs(t, err, ErrMissingRole)
		assert.Empty(t, token)
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("successful generation", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("user123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed
		claims, err := manager.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Empty(t, claims.Role) // Refresh tokens don't have roles
		assert.Equal(t, RefreshTokenType, claims.TokenType)
		assert.Equal(t, "test", claims.Issuer)
		assert.NotEmpty(t, claims.ID)
	})

	t.Run("missing user ID", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("")
		assert.ErrorIs(t, err, ErrMissingUserID)
		assert.Empty(t, token)
	})
}

func TestGenerateTokenPair(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("successful generation", func(t *testing.T) {
		pair, err := manager.GenerateTokenPair("user123", "volunteer")
		require.NoError(t, err)
		assert.NotNil(t, pair)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.Equal(t, int64(900), pair.ExpiresIn) // 15 minutes = 900 seconds

		// Verify both tokens
		accessClaims, err := manager.ValidateAccessToken(pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", accessClaims.UserID)
		assert.Equal(t, "volunteer", accessClaims.Role)

		refreshClaims, err := manager.ValidateRefreshToken(pair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", refreshClaims.UserID)
	})

	t.Run("missing user ID", func(t *testing.T) {
		pair, err := manager.GenerateTokenPair("", "volunteer")
		assert.Error(t, err)
		assert.Nil(t, pair)
	})

	t.Run("missing role", func(t *testing.T) {
		pair, err := manager.GenerateTokenPair("user123", "")
		assert.Error(t, err)
		assert.Nil(t, pair)
	})
}

func TestValidateAccessToken(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("valid token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		claims, err := manager.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "volunteer", claims.Role)
		assert.Equal(t, AccessTokenType, claims.TokenType)
	})

	t.Run("empty token", func(t *testing.T) {
		claims, err := manager.ValidateAccessToken("")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("malformed token", func(t *testing.T) {
		claims, err := manager.ValidateAccessToken("not-a-jwt-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("invalid signature", func(t *testing.T) {
		// Create token with different secret
		otherManager := NewManager(&Config{
			AccessSecret:       "different-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "test",
		})
		token, err := otherManager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		claims, err := manager.ValidateAccessToken(token)
		assert.ErrorIs(t, err, ErrInvalidSignature)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create manager with very short expiry
		expiredManager := NewManager(&Config{
			AccessSecret:       "test-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  1 * time.Millisecond,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "test",
		})

		token, err := expiredManager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		claims, err := manager.ValidateAccessToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("wrong token type (refresh token)", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		// When validating with wrong secret, we get signature error first
		claims, err := manager.ValidateAccessToken(token)
		assert.Error(t, err) // Will be signature error since refresh uses different secret
		assert.Nil(t, claims)
	})

	t.Run("missing role in access token", func(t *testing.T) {
		// Manually create a token without role
		now := time.Now()
		claims := Claims{
			UserID:    "user123",
			Role:      "", // Missing role
			TokenType: AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user123",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
				Issuer:    "test",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("test-secret"))
		require.NoError(t, err)

		validatedClaims, err := manager.ValidateAccessToken(signedToken)
		assert.ErrorIs(t, err, ErrMissingRole)
		assert.Nil(t, validatedClaims)
	})
}

func TestValidateRefreshToken(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("valid token", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		claims, err := manager.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, RefreshTokenType, claims.TokenType)
	})

	t.Run("wrong token type (access token)", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		// When validating with wrong secret, we get signature error first
		claims, err := manager.ValidateRefreshToken(token)
		assert.Error(t, err) // Will be signature error since access uses different secret
		assert.Nil(t, claims)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// Create manager with very short expiry
		expiredManager := NewManager(&Config{
			AccessSecret:       "test-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 1 * time.Millisecond,
			Issuer:             "test",
		})

		token, err := expiredManager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		claims, err := manager.ValidateRefreshToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Nil(t, claims)
	})
}

func TestRefreshTokenPair(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("successful refresh", func(t *testing.T) {
		// Generate initial token pair
		initialPair, err := manager.GenerateTokenPair("user123", "volunteer")
		require.NoError(t, err)

		// Refresh tokens
		newPair, oldTokenID, err := manager.RefreshTokenPair(initialPair.RefreshToken, "volunteer")
		require.NoError(t, err)
		assert.NotNil(t, newPair)
		assert.NotEmpty(t, oldTokenID)

		// Verify new tokens are different
		assert.NotEqual(t, initialPair.AccessToken, newPair.AccessToken)
		assert.NotEqual(t, initialPair.RefreshToken, newPair.RefreshToken)

		// Verify new tokens are valid
		accessClaims, err := manager.ValidateAccessToken(newPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", accessClaims.UserID)
		assert.Equal(t, "volunteer", accessClaims.Role)

		refreshClaims, err := manager.ValidateRefreshToken(newPair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", refreshClaims.UserID)

		// Verify old token ID can be used for blacklisting
		assert.NotEmpty(t, oldTokenID)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		newPair, oldTokenID, err := manager.RefreshTokenPair("invalid-token", "volunteer")
		assert.Error(t, err)
		assert.Nil(t, newPair)
		assert.Empty(t, oldTokenID)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// Create manager with very short expiry
		expiredManager := NewManager(&Config{
			AccessSecret:       "test-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 1 * time.Millisecond,
			Issuer:             "test",
		})

		token, err := expiredManager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		newPair, oldTokenID, err := manager.RefreshTokenPair(token, "volunteer")
		assert.Error(t, err)
		assert.Nil(t, newPair)
		assert.Empty(t, oldTokenID)
	})

	t.Run("using access token for refresh", func(t *testing.T) {
		accessToken, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		newPair, oldTokenID, err := manager.RefreshTokenPair(accessToken, "volunteer")
		assert.Error(t, err)
		assert.Nil(t, newPair)
		assert.Empty(t, oldTokenID)
	})
}

func TestParseToken(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("parse valid token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		claims, err := manager.ParseToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "volunteer", claims.Role)
		assert.Equal(t, AccessTokenType, claims.TokenType)
	})

	t.Run("parse expired token (still parses)", func(t *testing.T) {
		// Create manager with very short expiry
		expiredManager := NewManager(&Config{
			AccessSecret:       "test-secret",
			RefreshSecret:      "test-refresh-secret",
			AccessTokenExpiry:  1 * time.Millisecond,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "test",
		})

		token, err := expiredManager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		// ParseToken should still work for expired tokens
		claims, err := manager.ParseToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "volunteer", claims.Role)
	})

	t.Run("parse malformed token", func(t *testing.T) {
		claims, err := manager.ParseToken("not-a-jwt")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestGetTokenID(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("get token ID from valid token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		tokenID, err := manager.GetTokenID(token)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenID)

		// Verify it matches the ID from validation
		claims, err := manager.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, claims.ID, tokenID)
	})

	t.Run("get token ID from malformed token", func(t *testing.T) {
		tokenID, err := manager.GetTokenID("invalid-token")
		assert.Error(t, err)
		assert.Empty(t, tokenID)
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("IsExpiredError", func(t *testing.T) {
		assert.True(t, IsExpiredError(ErrExpiredToken))
		assert.True(t, IsExpiredError(jwt.ErrTokenExpired))
		assert.False(t, IsExpiredError(ErrInvalidToken))
		assert.False(t, IsExpiredError(nil))
	})

	t.Run("IsInvalidSignatureError", func(t *testing.T) {
		assert.True(t, IsInvalidSignatureError(ErrInvalidSignature))
		assert.True(t, IsInvalidSignatureError(jwt.ErrSignatureInvalid))
		assert.False(t, IsInvalidSignatureError(ErrInvalidToken))
		assert.False(t, IsInvalidSignatureError(nil))
	})

	t.Run("IsMalformedError", func(t *testing.T) {
		assert.True(t, IsMalformedError(ErrMalformedToken))
		assert.True(t, IsMalformedError(jwt.ErrTokenMalformed))
		assert.False(t, IsMalformedError(ErrInvalidToken))
		assert.False(t, IsMalformedError(nil))
	})
}

func TestTokenTypeValidation(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	t.Run("access token validated as access", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		claims, err := manager.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, AccessTokenType, claims.TokenType)
	})

	t.Run("refresh token validated as refresh", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		claims, err := manager.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, RefreshTokenType, claims.TokenType)
	})

	t.Run("cannot use refresh token as access token", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken("user123")
		require.NoError(t, err)

		// When validating with wrong secret, we get signature error first
		claims, err := manager.ValidateAccessToken(token)
		assert.Error(t, err) // Will be signature error since refresh uses different secret
		assert.Nil(t, claims)
	})

	t.Run("cannot use access token as refresh token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken("user123", "volunteer")
		require.NoError(t, err)

		// When validating with wrong secret, we get signature error first
		claims, err := manager.ValidateRefreshToken(token)
		assert.Error(t, err) // Will be signature error since access uses different secret
		assert.Nil(t, claims)
	})
}

func TestConcurrentTokenGeneration(t *testing.T) {
	manager := NewManager(&Config{
		AccessSecret:       "test-secret",
		RefreshSecret:      "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	})

	// Generate multiple tokens concurrently
	const numTokens = 100
	tokens := make([]string, numTokens)
	errors := make([]error, numTokens)

	done := make(chan bool, numTokens)

	for i := 0; i < numTokens; i++ {
		go func(index int) {
			token, err := manager.GenerateAccessToken("user123", "volunteer")
			tokens[index] = token
			errors[index] = err
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numTokens; i++ {
		<-done
	}

	// Verify all tokens were generated successfully and are unique
	seen := make(map[string]bool)
	for i := 0; i < numTokens; i++ {
		require.NoError(t, errors[i])
		assert.NotEmpty(t, tokens[i])
		assert.False(t, seen[tokens[i]], "duplicate token generated")
		seen[tokens[i]] = true
	}
}
