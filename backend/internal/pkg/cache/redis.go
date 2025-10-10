package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis cache configuration
type Config struct {
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

// DefaultConfig returns default Redis configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            "6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}
}

// Client represents a Redis client wrapper
type Client struct {
	redis  *redis.Client
	config *Config
}

// NewClient creates a new Redis client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        config.Password,
		DB:              config.DB,
		MaxRetries:      config.MaxRetries,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis: %s", addr)

	return &Client{
		redis:  rdb,
		config: config,
	}, nil
}

// Get retrieves a value from Redis by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s not found", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// GetJSON retrieves a JSON value from Redis and unmarshals it into the provided struct
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	// Retry logic for critical session retrieval operations
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		val, err := c.Get(ctx, key)
		if err == nil {
			// Successfully retrieved, now unmarshal
			if err := json.Unmarshal([]byte(val), dest); err != nil {
				return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
			}
			return nil
		}
		lastErr = err
		
		// Don't retry if key simply doesn't exist
		if err.Error() == fmt.Sprintf("key %s not found", key) {
			return err
		}
		
		// Exponential backoff: 10ms, 20ms, 40ms
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(10*(1<<uint(attempt))) * time.Millisecond)
		}
	}
	
	return fmt.Errorf("failed to get JSON after %d attempts: %w", maxRetries, lastErr)
}

// Set stores a value in Redis with an expiration time
// If expiration is 0, the key will never expire
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.redis.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// SetJSON marshals the provided struct to JSON and stores it in Redis
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	// Retry logic for critical session storage operations
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = c.Set(ctx, key, jsonData, expiration)
		if err == nil {
			return nil
		}
		lastErr = err
		
		// Exponential backoff: 10ms, 20ms, 40ms
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(10*(1<<uint(attempt))) * time.Millisecond)
		}
	}
	
	return fmt.Errorf("failed to set JSON after %d attempts: %w", maxRetries, lastErr)
}

// Delete removes one or more keys from Redis
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	// Retry logic for critical delete operations
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := c.redis.Del(ctx, keys...).Err()
		if err == nil {
			return nil
		}
		lastErr = err
		
		// Exponential backoff: 10ms, 20ms, 40ms
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(10*(1<<uint(attempt))) * time.Millisecond)
		}
	}
	
	return fmt.Errorf("failed to delete keys after %d attempts: %w", maxRetries, lastErr)
}

// Exists checks if a key exists in Redis
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if key %s exists: %w", key, err)
	}
	return count > 0, nil
}

// Expire sets a timeout on a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.redis.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration on key %s: %w", key, err)
	}
	return nil
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	duration, err := c.redis.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return duration, nil
}

// Increment increments the integer value of a key by one
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// IncrementBy increments the integer value of a key by the given amount
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := c.redis.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return val, nil
}

// SetNX sets a key only if it does not already exist
// Returns true if the key was set, false if it already existed
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	success, err := c.redis.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set key %s with NX: %w", key, err)
	}
	return success, nil
}

// FlushAll removes all keys from all databases (use with caution!)
func (c *Client) FlushAll(ctx context.Context) error {
	err := c.redis.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all databases: %w", err)
	}
	return nil
}

// FlushDB removes all keys from the current database
func (c *Client) FlushDB(ctx context.Context) error {
	err := c.redis.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}
	return nil
}

// Health checks the health of the Redis connection
func (c *Client) Health(ctx context.Context) error {
	if err := c.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	return nil
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	if err := c.redis.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	log.Println("Redis connection closed successfully")
	return nil
}

// GetClient returns the underlying Redis client for advanced operations
func (c *Client) GetClient() *redis.Client {
	return c.redis
}

// SessionStorage provides helper methods for session management
type SessionStorage struct {
	client        *Client
	sessionPrefix string
	defaultExpiry time.Duration
}

// NewSessionStorage creates a new session storage with the given Redis client
func NewSessionStorage(client *Client, sessionPrefix string, defaultExpiry time.Duration) *SessionStorage {
	if sessionPrefix == "" {
		sessionPrefix = "session:"
	}
	if defaultExpiry == 0 {
		defaultExpiry = 24 * time.Hour
	}

	return &SessionStorage{
		client:        client,
		sessionPrefix: sessionPrefix,
		defaultExpiry: defaultExpiry,
	}
}

// SetSession stores a session with the given ID and data
func (s *SessionStorage) SetSession(ctx context.Context, sessionID string, data interface{}) error {
	key := s.sessionPrefix + sessionID
	return s.client.SetJSON(ctx, key, data, s.defaultExpiry)
}

// GetSession retrieves a session by ID
func (s *SessionStorage) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := s.sessionPrefix + sessionID
	err := s.client.GetJSON(ctx, key, dest)
	if err != nil {
		// Log the error but don't expose internal details
		return fmt.Errorf("session not found or expired")
	}
	return nil
}

// DeleteSession removes a session by ID
func (s *SessionStorage) DeleteSession(ctx context.Context, sessionID string) error {
	key := s.sessionPrefix + sessionID
	return s.client.Delete(ctx, key)
}

// RefreshSession extends the TTL of a session
func (s *SessionStorage) RefreshSession(ctx context.Context, sessionID string) error {
	key := s.sessionPrefix + sessionID
	return s.client.Expire(ctx, key, s.defaultExpiry)
}

// SessionExists checks if a session exists
func (s *SessionStorage) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	key := s.sessionPrefix + sessionID
	return s.client.Exists(ctx, key)
}
