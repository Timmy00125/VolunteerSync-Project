package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Logger    LoggerConfig
	RateLimit RateLimitConfig
}

// AppConfig contains general application settings
type AppConfig struct {
	Env  string // development, staging, production
	Port string
	Mode string // debug, release
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	DB              int
	MaxRetries      int
	PoolSize        int
	MinIdleConns    int
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// JWTConfig contains JWT token settings
type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	AllowedOrigins []string
}

// LoggerConfig contains logging settings
type LoggerConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	WithCaller bool
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// Load reads configuration from environment variables
// Returns a Config struct with all settings populated
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "volunteersync"),
			Password:        getEnv("DB_PASSWORD", "volunteersync"),
			DBName:          getEnv("DB_NAME", "volunteersync"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "300s"),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", "600s"),
		},
		Redis: RedisConfig{
			Host:            getEnv("REDIS_HOST", "localhost"),
			Port:            getEnv("REDIS_PORT", "6379"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			DB:              getEnvAsInt("REDIS_DB", 0),
			MaxRetries:      getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:        getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
			ConnMaxIdleTime: getEnvAsDuration("REDIS_CONN_MAX_IDLE_TIME", "5m"),
			DialTimeout:     getEnvAsDuration("REDIS_DIAL_TIMEOUT", "5s"),
			ReadTimeout:     getEnvAsDuration("REDIS_READ_TIMEOUT", "3s"),
			WriteTimeout:    getEnvAsDuration("REDIS_WRITE_TIMEOUT", "3s"),
		},
		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-access-secret-change-in-production"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_TOKEN_EXPIRY", "15m"),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_TOKEN_EXPIRY", "168h"), // 7 days
			Issuer:             getEnv("JWT_ISSUER", "volunteersync"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{
				"http://localhost:3000",
				"http://localhost:8080",
			}),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			WithCaller: getEnvAsBool("LOG_WITH_CALLER", true),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsDuration("RATE_LIMIT_WINDOW", "1m"),
		},
	}

	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all critical configuration values are present and valid
func (c *Config) Validate() error {
	// Database validation
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Redis validation
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if c.Redis.Port == "" {
		return fmt.Errorf("REDIS_PORT is required")
	}

	// JWT validation
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if len(c.JWT.AccessSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	if len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 characters")
	}

	// App validation
	if c.App.Port == "" {
		return fmt.Errorf("APP_PORT is required")
	}

	// Logger validation
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logger.Level] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}

	validLogFormats := map[string]bool{"json": true, "text": true}
	if !validLogFormats[c.Logger.Format] {
		return fmt.Errorf("LOG_FORMAT must be one of: json, text")
	}

	return nil
}

// GetDatabaseDSN returns a PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// Helper functions for environment variable parsing

// getEnv retrieves an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer with a fallback default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool retrieves an environment variable as a boolean with a fallback default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration retrieves an environment variable as a time.Duration with a fallback default
func getEnvAsDuration(key, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}

	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// If parsing fails, try to parse the default
		duration, _ = time.ParseDuration(defaultValue)
	}

	return duration
}

// getEnvAsSlice retrieves an environment variable as a string slice (comma-separated)
// with a fallback default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Split by comma and trim spaces
	values := []string{}
	for _, v := range splitAndTrim(valueStr, ",") {
		if v != "" {
			values = append(values, v)
		}
	}

	if len(values) == 0 {
		return defaultValue
	}

	return values
}

// splitAndTrim splits a string by a delimiter and trims whitespace from each part
func splitAndTrim(s, delimiter string) []string {
	parts := []string{}
	for _, part := range splitString(s, delimiter) {
		trimmed := trimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}

// splitString splits a string by a delimiter
func splitString(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	current := ""

	for i := 0; i < len(s); i++ {
		if i+len(delimiter) <= len(s) && s[i:i+len(delimiter)] == delimiter {
			result = append(result, current)
			current = ""
			i += len(delimiter) - 1
		} else {
			current += string(s[i])
		}
	}

	result = append(result, current)
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && isWhitespace(s[start]) {
		start++
	}

	// Trim trailing whitespace
	for end > start && isWhitespace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isWhitespace checks if a byte is a whitespace character
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
