package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Common errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrMalformedToken   = errors.New("malformed token")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrMissingUserID    = errors.New("missing user ID in claims")
	ErrMissingRole      = errors.New("missing role in claims")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

// Config holds JWT configuration
type Config struct {
	AccessSecret       string
	RefreshSecret      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// DefaultConfig returns default JWT configuration
// Access tokens: 15 minutes
// Refresh tokens: 7 days
func DefaultConfig() *Config {
	return &Config{
		AccessSecret:       "change-this-secret-in-production",
		RefreshSecret:      "change-this-refresh-secret-in-production",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "volunteersync",
	}
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT token operations
type Manager struct {
	config *Config
}

// NewManager creates a new JWT manager with the given configuration
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config: config,
	}
}

// GenerateAccessToken creates a new access token for the given user
// Access tokens are short-lived (15 minutes by default) and contain user ID and role
func (m *Manager) GenerateAccessToken(userID, role string) (string, error) {
	if userID == "" {
		return "", ErrMissingUserID
	}
	if role == "" {
		return "", ErrMissingRole
	}

	now := time.Now()
	expiresAt := now.Add(m.config.AccessTokenExpiry)

	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    m.config.Issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.config.AccessSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken creates a new refresh token for the given user
// Refresh tokens are long-lived (7 days by default) and are used to obtain new access tokens
func (m *Manager) GenerateRefreshToken(userID string) (string, error) {
	if userID == "" {
		return "", ErrMissingUserID
	}

	now := time.Now()
	expiresAt := now.Add(m.config.RefreshTokenExpiry)

	claims := Claims{
		UserID:    userID,
		Role:      "", // Refresh tokens don't need role information
		TokenType: RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    m.config.Issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.config.RefreshSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

// GenerateTokenPair creates both access and refresh tokens for the given user
func (m *Manager) GenerateTokenPair(userID, role string) (*TokenPair, error) {
	accessToken, err := m.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := m.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.config.AccessTokenExpiry.Seconds()),
	}, nil
}

// ValidateAccessToken validates an access token and returns its claims
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.config.AccessSecret, AccessTokenType)
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.config.RefreshSecret, RefreshTokenType)
}

// validateToken is a helper function that validates a token with the given secret
func (m *Manager) validateToken(tokenString, secret string, expectedType TokenType) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrMalformedToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: expected %s, got %s", ErrInvalidTokenType, expectedType, claims.TokenType)
	}

	// Verify required fields
	if claims.UserID == "" {
		return nil, ErrMissingUserID
	}

	// Access tokens must have a role
	if expectedType == AccessTokenType && claims.Role == "" {
		return nil, ErrMissingRole
	}

	return claims, nil
}

// ParseToken parses a token without validating it (for debugging/testing purposes)
// This should NOT be used for authentication
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshTokenPair validates a refresh token and generates a new token pair
// The old refresh token should be invalidated by the caller (typically stored in Redis)
// Returns the new token pair and the old refresh token ID for invalidation
func (m *Manager) RefreshTokenPair(refreshTokenString, userRole string) (*TokenPair, string, error) {
	// Validate the refresh token
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	tokenPair, err := m.GenerateTokenPair(claims.UserID, userRole)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate new token pair: %w", err)
	}

	// Return the old token ID for invalidation
	oldTokenID := claims.ID

	return tokenPair, oldTokenID, nil
}

// GetTokenID extracts the token ID (jti claim) from a token without full validation
func (m *Manager) GetTokenID(tokenString string) (string, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// IsExpiredError checks if an error is due to token expiration
func IsExpiredError(err error) bool {
	return errors.Is(err, ErrExpiredToken) || errors.Is(err, jwt.ErrTokenExpired)
}

// IsInvalidSignatureError checks if an error is due to invalid signature
func IsInvalidSignatureError(err error) bool {
	return errors.Is(err, ErrInvalidSignature) || errors.Is(err, jwt.ErrSignatureInvalid)
}

// IsMalformedError checks if an error is due to malformed token
func IsMalformedError(err error) bool {
	return errors.Is(err, ErrMalformedToken) || errors.Is(err, jwt.ErrTokenMalformed)
}
