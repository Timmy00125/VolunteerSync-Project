# JWT Package

This package provides JWT (JSON Web Token) utilities for the VolunteerSync platform's authentication system.

## Features

- **Access Tokens**: Short-lived tokens (15 minutes) for API authentication
- **Refresh Tokens**: Long-lived tokens (7 days) for obtaining new access tokens
- **Token Rotation**: Secure refresh token rotation to prevent reuse attacks
- **Sliding Window Sessions**: Active users get session extended on each refresh (7 days from last activity)
- **Type Safety**: Separate validation for access and refresh tokens
- **Comprehensive Error Handling**: Detailed error types for all failure scenarios
- **Rate Limiting**: Token refresh is rate-limited (20 requests per 15 minutes)
- **Retry Logic**: Automatic retry for Redis operations (3 attempts with exponential backoff)

## Architecture

The package follows a clean, modular design:

- **Manager**: Core struct that handles all JWT operations
- **Config**: Configurable settings for token expiry, secrets, and issuer
- **Claims**: JWT claims structure with user ID, role, and token type
- **TokenPair**: Struct representing both access and refresh tokens

## Usage

### Initialize Manager

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/jwt"

// Use default configuration (development)
manager := jwt.NewManager(nil)

// Or use custom configuration (production)
config := &jwt.Config{
    AccessSecret:       "your-secure-access-secret",
    RefreshSecret:      "your-secure-refresh-secret",
    AccessTokenExpiry:  15 * time.Minute,
    RefreshTokenExpiry: 7 * 24 * time.Hour,
    Issuer:             "volunteersync",
}
manager := jwt.NewManager(config)
```

### Generate Token Pair (Login)

```go
// Generate both access and refresh tokens
tokenPair, err := manager.GenerateTokenPair(userID, userRole)
if err != nil {
    return err
}

// tokenPair.AccessToken - send in response body or header
// tokenPair.RefreshToken - store in httpOnly cookie
// tokenPair.ExpiresIn - access token expiry in seconds
```

### Validate Access Token (Protected Routes)

```go
// Extract token from Authorization header
tokenString := extractBearerToken(r.Header.Get("Authorization"))

// Validate the access token
claims, err := manager.ValidateAccessToken(tokenString)
if err != nil {
    if jwt.IsExpiredError(err) {
        return errors.New("token expired, please refresh")
    }
    return err
}

// Use claims
userID := claims.UserID
userRole := claims.Role
```

### Refresh Token Pair (Token Refresh Endpoint)

```go
// Extract refresh token from cookie
refreshToken := getRefreshTokenFromCookie(r)

// Validate and generate new token pair
newTokenPair, oldTokenID, err := manager.RefreshTokenPair(refreshToken, userRole)
if err != nil {
    return err
}

// Invalidate old refresh token in Redis
cache.Delete(ctx, "refresh_token:"+oldTokenID)

// Store new refresh token ID in Redis
cache.Set(ctx, "refresh_token:"+newTokenID, userID, 7*24*time.Hour)

// Return new token pair
```

## Security Features

### Token Rotation

Each token refresh generates a new access token AND a new refresh token. The old refresh token is invalidated, preventing token reuse attacks.

### Separate Secrets

Access tokens and refresh tokens use different signing secrets. This prevents an attacker from using an access token as a refresh token.

### Token Type Validation

The package enforces token type validation, ensuring access tokens can only be validated as access tokens and refresh tokens as refresh tokens.

### Expiry Validation

All tokens are validated for expiry. Expired tokens return `ErrExpiredToken`.

### Signature Validation

Tokens signed with incorrect secrets return `ErrInvalidSignature`.

## Error Handling

The package provides specific error types:

- `ErrInvalidToken`: Token is malformed or invalid
- `ErrExpiredToken`: Token has expired
- `ErrInvalidSignature`: Token signature verification failed
- `ErrMalformedToken`: Token format is incorrect
- `ErrInvalidClaims`: Claims structure is invalid
- `ErrMissingUserID`: User ID is missing from claims
- `ErrMissingRole`: Role is missing from access token claims
- `ErrInvalidTokenType`: Token type doesn't match expected type

Helper functions for error checking:

```go
if jwt.IsExpiredError(err) {
    // Handle expired token
}

if jwt.IsInvalidSignatureError(err) {
    // Handle invalid signature
}

if jwt.IsMalformedError(err) {
    // Handle malformed token
}
```

## Token Claims

### Access Token Claims

- `user_id`: User's unique identifier
- `role`: User's role (volunteer, coordinator, org_admin, super_admin)
- `token_type`: "access"
- `sub`: Subject (user ID)
- `iss`: Issuer (volunteersync)
- `iat`: Issued at timestamp
- `exp`: Expiration timestamp
- `jti`: JWT ID (unique token identifier)

### Refresh Token Claims

- `user_id`: User's unique identifier
- `token_type`: "refresh"
- `sub`: Subject (user ID)
- `iss`: Issuer (volunteersync)
- `iat`: Issued at timestamp
- `exp`: Expiration timestamp
- `jti`: JWT ID (unique token identifier)

Note: Refresh tokens do NOT include role information. Role must be fetched from the database during refresh.

## Integration with Redis

The JWT package is designed to work with the Redis cache package for:

1. **Refresh Token Storage**: Store refresh token IDs in Redis with user ID as value
2. **Token Blacklist**: Invalidate refresh tokens immediately on logout
3. **Refresh Token Rotation**: Track old token IDs for invalidation

Example Redis integration:

```go
// Store refresh token
tokenID, _ := manager.GetTokenID(refreshToken)
cache.Set(ctx, "refresh_token:"+tokenID, userID, 7*24*time.Hour)

// Check if refresh token is valid (not blacklisted)
exists, _ := cache.Exists(ctx, "refresh_token:"+tokenID)
if !exists {
    return errors.New("refresh token has been revoked")
}

// Revoke refresh token on logout
cache.Delete(ctx, "refresh_token:"+tokenID)
```

## Testing

The package includes comprehensive unit tests covering:

- Token generation (access, refresh, pairs)
- Token validation (valid, expired, invalid signature, malformed)
- Token rotation (successful, failed, with invalid tokens)
- Concurrent token generation
- Error handling and helper functions
- Token type validation

Run tests:

```bash
go test ./internal/pkg/jwt/... -v
```

## Production Considerations

1. **Use Strong Secrets**: Generate cryptographically secure random secrets
2. **Environment Variables**: Never hardcode secrets in source code
3. **HTTPS Only**: Always use HTTPS to protect tokens in transit
4. **httpOnly Cookies**: Store refresh tokens in httpOnly, secure, SameSite cookies
5. **Access Token Storage**: Store access tokens in memory (not localStorage due to XSS risk)
6. **Token Rotation**: Always invalidate old refresh tokens in Redis after rotation
7. **Rate Limiting**: Implement rate limiting on token refresh endpoints
8. **Audit Logging**: Log all token generation and validation failures

## Future Enhancements

- [ ] Support for token revocation lists (blacklist expired tokens)
- [ ] Support for multiple device sessions per user
- [ ] Token rotation detection (detect reuse of old refresh tokens)
- [ ] Admin endpoints for force logout (revoke all user tokens)
- [ ] Key rotation support (rotate signing secrets)
