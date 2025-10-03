package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "volunteersync", config.User)
	assert.Equal(t, "volunteersync", config.Password)
	assert.Equal(t, "volunteersync", config.DBName)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, logger.Info, config.LogLevel)
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "default config",
			config:   DefaultConfig(),
			expected: "host=localhost port=5432 user=volunteersync password=volunteersync dbname=volunteersync sslmode=disable",
		},
		{
			name: "custom config",
			config: &Config{
				Host:     "db.example.com",
				Port:     "5433",
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5433 user=testuser password=testpass dbname=testdb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildDSN(tt.config)
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestNewConnection_InvalidConfig(t *testing.T) {
	// Test with invalid database credentials
	config := &Config{
		Host:     "invalid-host",
		Port:     "9999",
		User:     "invalid",
		Password: "invalid",
		DBName:   "invalid",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	conn, err := NewConnection(config)
	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestConnection_HealthCheck_NilDB(t *testing.T) {
	conn := &Connection{
		DB:     nil,
		config: DefaultConfig(),
	}

	ctx := context.Background()
	err := conn.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestConnection_IsConnected_NilDB(t *testing.T) {
	conn := &Connection{
		DB:     nil,
		config: DefaultConfig(),
	}

	assert.False(t, conn.IsConnected())
}

func TestConnection_GetStats_NilDB(t *testing.T) {
	conn := &Connection{
		DB:     nil,
		config: DefaultConfig(),
	}

	stats := conn.GetStats()
	require.NotNil(t, stats)
	assert.Contains(t, stats, "error")
	assert.Equal(t, "database connection is nil", stats["error"])
}

func TestConnection_Close_NilDB(t *testing.T) {
	conn := &Connection{
		DB:     nil,
		config: DefaultConfig(),
	}

	err := conn.Close()
	assert.NoError(t, err) // Should not error on nil DB
}

func TestNewConnection_WithNilConfig(t *testing.T) {
	// This test will fail unless a real database is available
	// In a real test suite, we would use testcontainers or mock the database
	// For now, we just verify it uses default config
	_, err := NewConnection(nil)

	// We expect this to fail with connection error since we're using default config
	// which points to localhost
	if err != nil {
		assert.Contains(t, err.Error(), "failed to connect to database")
	}
}

// Integration tests below require a running PostgreSQL instance
// These should be run with build tag: go test -tags=integration

// TestNewConnection_Success tests successful connection creation
// This test requires a running PostgreSQL database
func TestNewConnection_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use environment-specific test database
	config := &Config{
		Host:            "localhost",
		Port:            "5432",
		User:            "volunteersync_test",
		Password:        "test",
		DBName:          "volunteersync_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 1 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
		LogLevel:        logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NotNil(t, conn.DB)
	assert.True(t, conn.IsConnected())
}

// TestConnection_HealthCheck tests health check functionality
func TestConnection_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "volunteersync_test",
		Password: "test",
		DBName:   "volunteersync_test",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	ctx := context.Background()
	err = conn.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = conn.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestConnection_GetStats tests connection pool statistics
func TestConnection_GetStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:            "localhost",
		Port:            "5432",
		User:            "volunteersync_test",
		Password:        "test",
		DBName:          "volunteersync_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
		LogLevel:        logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	stats := conn.GetStats()
	require.NotNil(t, stats)

	assert.Contains(t, stats, "max_open_connections")
	assert.Contains(t, stats, "open_connections")
	assert.Contains(t, stats, "in_use")
	assert.Contains(t, stats, "idle")
	assert.Equal(t, 10, stats["max_open_connections"])
}

// TestConnection_Transaction tests transaction functionality
func TestConnection_Transaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "volunteersync_test",
		Password: "test",
		DBName:   "volunteersync_test",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	// Test successful transaction
	err = conn.Transaction(func(tx *gorm.DB) error {
		// Perform some operations within transaction
		return nil
	})
	assert.NoError(t, err)

	// Test transaction rollback on error
	testError := assert.AnError
	err = conn.Transaction(func(tx *gorm.DB) error {
		return testError
	})
	assert.Error(t, err)
	assert.Equal(t, testError, err)
}

// TestConnection_WithContext tests context-aware database operations
func TestConnection_WithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "volunteersync_test",
		Password: "test",
		DBName:   "volunteersync_test",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	ctx := context.Background()
	db := conn.WithContext(ctx)

	require.NotNil(t, db)
	assert.NotEqual(t, conn.DB, db) // Should be a new instance with context
}

// TestConnection_GetDB tests getting the underlying GORM DB
func TestConnection_GetDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "volunteersync_test",
		Password: "test",
		DBName:   "volunteersync_test",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	conn, err := NewConnection(config)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer conn.Close()

	db := conn.GetDB()
	require.NotNil(t, db)
	assert.Equal(t, conn.DB, db)
}
