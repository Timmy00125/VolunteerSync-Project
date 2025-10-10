# Session Management

This document describes the session management strategy for VolunteerSync authentication.

## Overview

VolunteerSync uses a hybrid approach combining stateless JWT tokens with stateful session tracking in Redis. This provides the security benefits of session-based auth with the scalability of JWT.

## Architecture

### Components

1. **JWT Access Token** (15 minutes)
   - Short-lived, stateless
   - Contains user ID and role
   - Used for authenticating API requests
   - No database lookup required

2. **JWT Refresh Token** (7 days)
   - Long-lived, tracked in Redis
   - Used to obtain new access/refresh token pairs
   - Each has a unique token ID (JTI claim)

3. **Redis Session Store**
   - Maps token ID → session data (user ID, role, expiry)
   - TTL matches JWT refresh token expiry (7 days)
   - Enables immediate token revocation

## Token Refresh Flow

```
Client                    Server                     Redis
  |                          |                          |
  |-- POST /auth/refresh --->|                          |
  |    (refresh_token)        |                          |
  |                          |-- Get session by ID ---->|
  |                          |<-- Session data ---------|
  |                          |                          |
  |                          |-- Validate JWT --------->|
  |                          |                          |
  |                          |-- Delete old session --->|
  |                          |                          |
  |                          |-- Generate new tokens -->|
  |                          |                          |
  |                          |-- Store new session ---->|
  |                          |                          |
  |<-- New token pair -------|                          |
```

## Key Security Features

### 1. Token Rotation
Every refresh generates:
- New access token (15 min expiry)
- New refresh token (7 day expiry)
- Old refresh token is invalidated

### 2. Reuse Attack Prevention
- Old session deleted BEFORE new session created
- If attacker tries to use old refresh token → fails (session not found)
- If legitimate user's token stolen → both get logged out on next use

### 3. Rate Limiting
- Refresh endpoint: 20 requests per 15 minutes per IP
- Login endpoint: 5 requests per 15 minutes per IP
- Registration: 5 requests per 15 minutes per IP

### 4. Sliding Window Sessions
- Each successful refresh extends session TTL to 7 days
- Active users stay logged in indefinitely
- Inactive users logged out after 7 days

### 5. Retry Logic
Redis operations retry up to 3 times with exponential backoff:
- Attempt 1: immediate
- Attempt 2: +10ms
- Attempt 3: +20ms
- Attempt 4: +40ms

Handles transient Redis connection issues gracefully.

## Session Data Structure

```go
type refreshSessionData struct {
    UserID    string    `json:"user_id"`     // UUID string
    UserRole  string    `json:"user_role"`   // "volunteer", "org_admin", etc.
    TokenID   string    `json:"token_id"`    // JWT ID (jti claim)
    ExpiresAt time.Time `json:"expires_at"`  // Must match JWT expiry
}
```

Stored in Redis with key: `auth:refresh:{token_id}`

## Error Handling

### Token Validation Errors
- `ErrExpiredToken` → "Token has expired" (401)
- `ErrInvalidSignature` → "Invalid token signature" (401)
- `ErrInvalidToken` → "Invalid token" (401)
- `ErrMissingUserID` → "Invalid token claims" (401)

### Session Errors
- Session not found → "Refresh token has been revoked" (401)
- Redis unavailable → Retries 3x, then "Failed to persist session" (500)

### Rate Limit Errors
- Too many requests → "Too many refresh attempts" (429)
- Includes `Retry-After` header with seconds to wait

## Best Practices

### Client Implementation
```javascript
// Store access token in memory (not localStorage - XSS risk)
let accessToken = response.access_token;

// Store refresh token in httpOnly cookie (backend sets this)
// OR in secure storage if mobile app

// On 401 from API:
if (error.status === 401) {
  // Try to refresh
  const newTokens = await refreshToken();
  if (newTokens) {
    accessToken = newTokens.access_token;
    // Retry original request with new token
  } else {
    // Redirect to login
  }
}
```

### Token Expiry Handling
- Access token expires: Client should automatically refresh
- Refresh token expires: User must re-login (after 7 days inactive)
- Both tokens compromised: Attacker and user both logged out on next refresh

## Monitoring & Logging

### Logged Events
- **Register**: User ID, success/failure
- **Login**: User ID, success/failure
- **Refresh**: User ID, old token ID, success
- **Logout**: User ID, success

### Metrics to Monitor
- Refresh rate (should be ~every 15 min per active user)
- Failed refresh attempts (indicates attack or bugs)
- Session creation/deletion rate
- Redis connection errors

## Migration & Deployment

### Rolling Updates
Safe to deploy with zero downtime:
1. Old servers continue working with existing sessions
2. New servers use improved session management
3. Sessions naturally migrate to new format on next refresh

### Redis Maintenance
- Sessions can be regenerated by having users re-login
- Redis backup not critical (sessions are temporary)
- Redis cluster recommended for high availability

## Performance Characteristics

### Access Token Validation
- **Latency**: <1ms (no database lookup)
- **Scalability**: Horizontal (stateless)

### Refresh Token Validation
- **Latency**: ~2-5ms (Redis lookup + JWT validation)
- **Scalability**: Limited by Redis (but very high throughput)
- **Retry overhead**: +70ms worst case (3 attempts)

### Memory Usage
- Per session: ~200 bytes in Redis
- 10,000 active users: ~2MB in Redis

## Future Enhancements

### Completed ✅
- [x] Token rotation on refresh
- [x] Session tracking in Redis
- [x] Rate limiting on auth endpoints
- [x] Retry logic for Redis operations
- [x] Sliding window sessions
- [x] Comprehensive error logging

### Planned
- [ ] Multi-device session management (track all devices)
- [ ] Force logout all devices endpoint
- [ ] Session activity tracking (last used timestamp)
- [ ] Suspicious activity detection
- [ ] Key rotation support (rotate JWT secrets)

## Troubleshooting

### "Session invalidated easily"
**Fixed** ✅
- Sessions now persist correctly with proper TTL
- Sliding window keeps active users logged in
- Retry logic handles transient Redis issues

### "Token has been revoked" error
Possible causes:
1. User logged out on another device
2. Token refresh race condition (multiple tabs)
3. Admin revoked user access
4. Redis session expired (after 7 days inactive)

Solution: User must re-login

### Rate limit errors
User sending too many refresh requests:
1. Check client is only refreshing when needed (not on every request)
2. Implement proper token expiry checking in client
3. Don't refresh token unless access token expired

### Redis connection errors
System retries 3 times automatically:
1. Check Redis health: `redis-cli ping`
2. Check network connectivity
3. Review Redis logs for errors
4. Consider Redis cluster for HA

## References

- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OAuth 2.0 Token Revocation](https://datatracker.ietf.org/doc/html/rfc7009)
