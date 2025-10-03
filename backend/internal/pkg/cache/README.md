# Redis Cache Package

This package provides Redis caching utilities for the VolunteerSync backend application.

## Features

- **Connection Management**: Configurable Redis client with connection pooling
- **Basic Operations**: Get, Set, Delete, Exists operations
- **JSON Support**: Automatic marshaling/unmarshaling of JSON data
- **TTL Management**: Support for key expiration and TTL queries
- **Atomic Operations**: Increment operations and SetNX (set if not exists)
- **Session Storage**: Built-in session management helpers
- **Health Checks**: Connection health verification

## Installation

The Redis client is already included as a dependency in `go.mod`:

```bash
go get github.com/redis/go-redis/v9
```

## Usage

### Basic Client Setup

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/cache"

// Use default configuration
client, err := cache.NewClient(nil)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Or use custom configuration
config := &cache.Config{
    Host:     "localhost",
    Port:     "6379",
    Password: "your_password",
    DB:       0,
}
client, err := cache.NewClient(config)
```

### Basic Operations

```go
ctx := context.Background()

// Set a key with expiration
err := client.Set(ctx, "user:123", "John Doe", 1*time.Hour)

// Get a key
value, err := client.Get(ctx, "user:123")

// Delete keys
err := client.Delete(ctx, "user:123", "user:456")

// Check if key exists
exists, err := client.Exists(ctx, "user:123")

// Set expiration
err := client.Expire(ctx, "user:123", 5*time.Minute)

// Get TTL
ttl, err := client.TTL(ctx, "user:123")
```

### JSON Operations

```go
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Store JSON
user := User{ID: "123", Name: "John Doe", Email: "john@example.com"}
err := client.SetJSON(ctx, "user:123", user, 1*time.Hour)

// Retrieve JSON
var retrievedUser User
err := client.GetJSON(ctx, "user:123", &retrievedUser)
```

### Atomic Operations

```go
// Increment counter
count, err := client.Increment(ctx, "page:views")

// Increment by value
count, err := client.IncrementBy(ctx, "page:views", 10)

// Set only if not exists (distributed lock)
success, err := client.SetNX(ctx, "lock:resource", "owner1", 30*time.Second)
if success {
    // Lock acquired
    defer client.Delete(ctx, "lock:resource")
    // ... perform work ...
}
```

### Session Management

```go
// Create session storage
sessionStorage := cache.NewSessionStorage(client, "session:", 24*time.Hour)

type SessionData struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

// Store session
data := SessionData{UserID: "123", Username: "john", Role: "admin"}
err := sessionStorage.SetSession(ctx, "session_abc123", data)

// Retrieve session
var session SessionData
err := sessionStorage.GetSession(ctx, "session_abc123", &session)

// Check if session exists
exists, err := sessionStorage.SessionExists(ctx, "session_abc123")

// Refresh session (extend TTL)
err := sessionStorage.RefreshSession(ctx, "session_abc123")

// Delete session
err := sessionStorage.DeleteSession(ctx, "session_abc123")
```

## Use Cases

### 1. JWT Token Blacklist (Logout)

```go
// When user logs out, blacklist the token
tokenID := "jwt_id_xyz"
expiry := 15 * time.Minute // Match token expiry
err := client.Set(ctx, "blacklist:"+tokenID, "1", expiry)

// Check if token is blacklisted
isBlacklisted, _ := client.Exists(ctx, "blacklist:"+tokenID)
```

### 2. Rate Limiting

```go
// Track login attempts per IP
ip := "192.168.1.1"
key := "ratelimit:login:" + ip

// Increment attempt count
attempts, err := client.Increment(ctx, key)
if attempts == 1 {
    // First attempt, set expiration
    client.Expire(ctx, key, 15*time.Minute)
}

if attempts > 5 {
    // Rate limit exceeded
    return errors.New("too many login attempts")
}
```

### 3. Geocoding Cache

```go
// Cache geocoding results to avoid API calls
address := "123 Main St, Springfield"
cacheKey := "geocode:" + address

// Try to get from cache
var location struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

err := client.GetJSON(ctx, cacheKey, &location)
if err != nil {
    // Not in cache, fetch from API
    location = geocodeAPI(address)
    // Store in cache for 30 days
    client.SetJSON(ctx, cacheKey, location, 30*24*time.Hour)
}
```

### 4. Organization/Opportunity Caching

```go
// Cache frequently accessed organizations
orgID := "org_123"
cacheKey := "org:" + orgID

var org Organization
err := client.GetJSON(ctx, cacheKey, &org)
if err != nil {
    // Not in cache, fetch from database
    org = fetchOrgFromDB(orgID)
    // Cache for 10 minutes
    client.SetJSON(ctx, cacheKey, org, 10*time.Minute)
}
```

### 5. Real-time Analytics

```go
// Track event registrations in real-time
eventID := "event_456"

// Increment registration count
count, err := client.Increment(ctx, "event:"+eventID+":registrations")

// Get current count
value, err := client.Get(ctx, "event:"+eventID+":registrations")
```

## Configuration

### Default Configuration

```go
Host:            "localhost"
Port:            "6379"
Password:        "" (no password)
DB:              0
MaxRetries:      3
PoolSize:        10
MinIdleConns:    2
ConnMaxIdleTime: 5 minutes
DialTimeout:     5 seconds
ReadTimeout:     3 seconds
WriteTimeout:    3 seconds
```

### Environment Variables

For production, configure via environment variables:

```bash
REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=secret
REDIS_DB=0
```

## Testing

### Run Unit Tests (Short Mode)

```bash
cd backend
go test ./internal/pkg/cache/... -short -v
```

### Run Integration Tests (Requires Redis)

```bash
# Start Redis with Docker
docker run -d -p 6379:6379 redis:7-alpine

# Run all tests
go test ./internal/pkg/cache/... -v
```

### Run Specific Test

```bash
go test ./internal/pkg/cache/... -run TestSessionStorage -v
```

## Best Practices

1. **Always use context**: Pass context to all operations for proper cancellation and timeout handling
2. **Set appropriate TTLs**: Don't let cache grow indefinitely
3. **Handle errors**: Redis operations can fail; always check errors
4. **Use prefixes**: Namespace your keys with prefixes (e.g., `user:`, `session:`, `cache:`)
5. **Connection pooling**: Reuse the client instance across your application
6. **Graceful shutdown**: Always call `client.Close()` on application shutdown

## Performance Considerations

- **Connection pooling**: Default pool size is 10, adjust based on load
- **Key expiration**: Always set TTL to prevent memory leaks
- **Batch operations**: Use pipelines for multiple operations (use `GetClient()` for advanced operations)
- **Serialization**: JSON operations have overhead; use plain strings for simple values

## Error Handling

```go
val, err := client.Get(ctx, "key")
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // Key doesn't exist - this is expected
        // Fetch from database or return default
    } else {
        // Actual error - log and handle
        log.Printf("Redis error: %v", err)
    }
}
```

## Health Checks

```go
// Check Redis health in your /health endpoint
err := client.Health(ctx)
if err != nil {
    // Redis is down
    log.Printf("Redis health check failed: %v", err)
}
```

## Advanced Usage

For advanced Redis operations not covered by the wrapper, access the underlying client:

```go
redisClient := client.GetClient()

// Use raw Redis commands
pipe := redisClient.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
_, err := pipe.Exec(ctx)
```

## Dependencies

- **github.com/redis/go-redis/v9**: Official Go Redis client

## Related Documentation

- [Redis Commands](https://redis.io/commands/)
- [go-redis Documentation](https://redis.uptrace.dev/)
- [Redis Best Practices](https://redis.io/docs/management/optimization/)
