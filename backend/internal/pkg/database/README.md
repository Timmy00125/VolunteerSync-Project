# Database Package

This package provides database connection utilities for the VolunteerSync backend application.

## Features

- **PostgreSQL Connection**: GORM-based PostgreSQL connection with configurable settings
- **Connection Pooling**: Configurable connection pool with max open/idle connections and lifetimes
- **Health Checks**: Built-in health check functionality for monitoring
- **Transaction Support**: Helper methods for database transactions
- **Context Support**: Context-aware database operations
- **Connection Statistics**: Real-time connection pool statistics
- **Graceful Shutdown**: Proper connection cleanup

## Usage

### Basic Connection

```go
package main

import (
    "log"
    "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/database"
)

func main() {
    // Create connection with default configuration
    conn, err := database.NewConnection(nil)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer conn.Close()

    // Use the connection
    db := conn.GetDB()
    // ... perform database operations
}
```

### Custom Configuration

```go
config := &database.Config{
    Host:            "db.example.com",
    Port:            "5432",
    User:            "myuser",
    Password:        "mypassword",
    DBName:          "mydb",
    SSLMode:         "require",
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 10 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
    LogLevel:        logger.Info,
}

conn, err := database.NewConnection(config)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()
```

### Environment-Based Configuration

```go
import "os"

config := &database.Config{
    Host:     os.Getenv("DB_HOST"),
    Port:     os.Getenv("DB_PORT"),
    User:     os.Getenv("DB_USER"),
    Password: os.Getenv("DB_PASSWORD"),
    DBName:   os.Getenv("DB_NAME"),
    SSLMode:  os.Getenv("DB_SSLMODE"),
}

conn, err := database.NewConnection(config)
```

### Health Checks

```go
ctx := context.Background()

if err := conn.HealthCheck(ctx); err != nil {
    log.Printf("Health check failed: %v", err)
}

// Check connection status
if !conn.IsConnected() {
    log.Println("Database connection is down")
}
```

### Connection Statistics

```go
stats := conn.GetStats()
log.Printf("Database stats: %+v", stats)

// Example output:
// {
//   "max_open_connections": 25,
//   "open_connections": 5,
//   "in_use": 2,
//   "idle": 3,
//   "wait_count": 0,
//   "wait_duration": "0s",
//   "max_idle_closed": 0,
//   "max_lifetime_closed": 0
// }
```

### Transactions

```go
err := conn.Transaction(func(tx *gorm.DB) error {
    // Create user
    user := &User{Email: "test@example.com"}
    if err := tx.Create(user).Error; err != nil {
        return err // Transaction will be rolled back
    }

    // Create profile
    profile := &Profile{UserID: user.ID}
    if err := tx.Create(profile).Error; err != nil {
        return err // Transaction will be rolled back
    }

    return nil // Transaction will be committed
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

### Context-Aware Operations

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

db := conn.WithContext(ctx)

var users []User
if err := db.Find(&users).Error; err != nil {
    log.Printf("Query failed: %v", err)
}
```

## Configuration Options

| Option            | Type            | Default         | Description                                  |
| ----------------- | --------------- | --------------- | -------------------------------------------- |
| `Host`            | string          | "localhost"     | Database host                                |
| `Port`            | string          | "5432"          | Database port                                |
| `User`            | string          | "volunteersync" | Database user                                |
| `Password`        | string          | "volunteersync" | Database password                            |
| `DBName`          | string          | "volunteersync" | Database name                                |
| `SSLMode`         | string          | "disable"       | SSL mode (disable, require, verify-ca, etc.) |
| `MaxOpenConns`    | int             | 25              | Maximum number of open connections           |
| `MaxIdleConns`    | int             | 5               | Maximum number of idle connections           |
| `ConnMaxLifetime` | time.Duration   | 5 minutes       | Maximum lifetime of a connection             |
| `ConnMaxIdleTime` | time.Duration   | 10 minutes      | Maximum time a connection can be idle        |
| `LogLevel`        | logger.LogLevel | logger.Info     | GORM log level (Silent, Error, Warn, Info)   |

## Testing

The package includes comprehensive unit and integration tests.

```bash
# Run all tests (skips integration tests without database)
go test ./internal/pkg/database/

# Run with verbose output
go test ./internal/pkg/database/ -v

# Run only unit tests (fast)
go test ./internal/pkg/database/ -short

# Run with coverage
go test ./internal/pkg/database/ -cover
```

### Integration Tests

Integration tests require a running PostgreSQL database with the following configuration:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=volunteersync_test
DB_PASSWORD=test
DB_NAME=volunteersync_test
```

Integration tests will automatically skip if the database is not available.

## Best Practices

1. **Use Default Configuration in Development**: The default configuration works with Docker Compose setup
2. **Environment Variables in Production**: Always use environment variables for production credentials
3. **Connection Pooling**: Adjust pool settings based on your application's concurrency needs
4. **Health Checks**: Implement health check endpoints using `HealthCheck()` method
5. **Graceful Shutdown**: Always call `Close()` when shutting down the application
6. **Transactions**: Use `Transaction()` helper for multi-operation atomic updates
7. **Context Awareness**: Use `WithContext()` for timeout-based operations
8. **Monitor Statistics**: Periodically check `GetStats()` to monitor connection pool health

## Thread Safety

All methods in the `Connection` type are thread-safe and can be called concurrently from multiple goroutines.

## Error Handling

The package returns descriptive errors wrapped with context:

- Connection failures: `"failed to connect to database: ..."`
- Ping failures: `"database ping failed: ..."`
- Transaction failures: Original error from transaction function

Always check and handle errors appropriately in your application.

## Performance Considerations

- **Prepared Statements**: Enabled by default for better performance
- **Skip Default Transactions**: GORM default transactions are disabled for better performance
- **Connection Reuse**: Connection pooling ensures efficient connection reuse
- **Context Timeouts**: Use context timeouts to prevent long-running queries

## License

See the main project LICENSE file.
