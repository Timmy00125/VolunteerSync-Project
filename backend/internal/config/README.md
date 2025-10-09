# Configuration Management

This package provides centralized configuration management for the VolunteerSync backend API.

## Overview

The `config` package loads all application configuration from environment variables and provides a structured, type-safe way to access configuration throughout the application.

## Usage

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/config"

func main() {
    // Load configuration from environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Access configuration values
    dbHost := cfg.Database.Host
    jwtSecret := cfg.JWT.AccessSecret
    appPort := cfg.App.Port
}
```

## Configuration Sections

### App Configuration

- `APP_ENV` - Application environment (development, staging, production)
- `APP_PORT` - HTTP server port (default: 8080)
- `GIN_MODE` - Gin framework mode (debug, release)

### Database Configuration

- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username (default: volunteersync)
- `DB_PASSWORD` - PostgreSQL password (default: volunteersync)
- `DB_NAME` - PostgreSQL database name (default: volunteersync)
- `DB_SSLMODE` - PostgreSQL SSL mode (default: disable)
- `DB_MAX_OPEN_CONNS` - Maximum open database connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Maximum idle database connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Maximum connection lifetime (default: 300s)
- `DB_CONN_MAX_IDLE_TIME` - Maximum connection idle time (default: 600s)

### Redis Configuration

- `REDIS_HOST` - Redis host (default: localhost)
- `REDIS_PORT` - Redis port (default: 6379)
- `REDIS_PASSWORD` - Redis password (optional)
- `REDIS_DB` - Redis database number (default: 0)
- `REDIS_MAX_RETRIES` - Maximum retry attempts (default: 3)
- `REDIS_POOL_SIZE` - Connection pool size (default: 10)
- `REDIS_MIN_IDLE_CONNS` - Minimum idle connections (default: 2)
- `REDIS_CONN_MAX_IDLE_TIME` - Maximum connection idle time (default: 5m)
- `REDIS_DIAL_TIMEOUT` - Dial timeout (default: 5s)
- `REDIS_READ_TIMEOUT` - Read timeout (default: 3s)
- `REDIS_WRITE_TIMEOUT` - Write timeout (default: 3s)

### JWT Configuration

- `JWT_ACCESS_SECRET` - Access token secret (minimum 32 characters, **required in production**)
- `JWT_REFRESH_SECRET` - Refresh token secret (minimum 32 characters, **required in production**)
- `JWT_ACCESS_TOKEN_EXPIRY` - Access token expiry duration (default: 15m)
- `JWT_REFRESH_TOKEN_EXPIRY` - Refresh token expiry duration (default: 168h / 7 days)
- `JWT_ISSUER` - JWT issuer claim (default: volunteersync)

### CORS Configuration

- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins (default: http://localhost:3000,http://localhost:8080)

### Logger Configuration

- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: info)
- `LOG_FORMAT` - Log format: json, text (default: json)
- `LOG_WITH_CALLER` - Include caller information in logs (default: true)

### Rate Limiting Configuration

- `RATE_LIMIT_REQUESTS` - Maximum requests per window (default: 100)
- `RATE_LIMIT_WINDOW` - Rate limit time window (default: 1m)

## Validation

The `Load()` function automatically validates all critical configuration values:

- Required fields (database host, port, user, etc.)
- JWT secrets must be at least 32 characters in production
- Valid log levels and formats
- Proper duration formats

If validation fails, `Load()` returns an error and the application should exit.

## Helper Methods

### GetDatabaseDSN()

Returns a PostgreSQL connection string:

```go
dsn := cfg.GetDatabaseDSN()
// "host=localhost port=5432 user=volunteersync password=... dbname=volunteersync sslmode=disable"
```

### GetRedisAddr()

Returns the Redis connection address:

```go
addr := cfg.GetRedisAddr()
// "localhost:6379"
```

## Security Notes

1. **Never commit secrets to version control** - use environment variables
2. **Use strong secrets in production** - minimum 32 characters for JWT secrets
3. **Set appropriate CORS origins** - don't use wildcards in production
4. **Enable SSL for database** - set `DB_SSLMODE=require` in production
5. **Use Redis password** - always set `REDIS_PASSWORD` in production

## Environment File Example

Create a `.env` file for local development (DO NOT commit this file):

```env
# Application
APP_ENV=development
APP_PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=volunteersync
DB_PASSWORD=volunteersync_dev_password
DB_NAME=volunteersync
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=volunteersync_redis_password

# JWT (generate secure random strings for production!)
JWT_ACCESS_SECRET=dev_jwt_access_secret_minimum_32_chars_change_in_production
JWT_REFRESH_SECRET=dev_jwt_refresh_secret_minimum_32_chars_change_in_production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

## Design Principles

1. **Single Responsibility** - Configuration loading and validation only
2. **DRY** - All environment variable parsing in one place
3. **Type Safety** - Strong typing for all configuration values
4. **Validation** - Early validation with clear error messages
5. **Defaults** - Sensible defaults for development
6. **Documentation** - Clear documentation for all settings
