package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
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
	LogLevel        logger.LogLevel
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            "5432",
		User:            "volunteersync",
		Password:        "volunteersync",
		DBName:          "volunteersync",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        logger.Info,
	}
}

// Connection represents a database connection wrapper
type Connection struct {
	DB     *gorm.DB
	config *Config
}

// NewConnection creates a new database connection with the given configuration
func NewConnection(config *Config) (*Connection, error) {
	if config == nil {
		config = DefaultConfig()
	}

	dsn := buildDSN(config)

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:            true, // Prepare statements for better performance
		DisableAutomaticPing:   false,
		SkipDefaultTransaction: true, // Disable default transaction for better performance
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database: %s@%s:%s/%s",
		config.User, config.Host, config.Port, config.DBName)

	return &Connection{
		DB:     db,
		config: config,
	}, nil
}

// buildDSN constructs a PostgreSQL connection string from config
func buildDSN(config *Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB == nil {
		return nil
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

// HealthCheck performs a health check on the database connection
func (c *Connection) HealthCheck(ctx context.Context) error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Ping with timeout
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection pool statistics
func (c *Connection) GetStats() map[string]interface{} {
	if c.DB == nil {
		return map[string]interface{}{
			"error": "database connection is nil",
		}
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (c *Connection) Transaction(fn func(tx *gorm.DB) error) error {
	return c.DB.Transaction(fn)
}

// WithContext returns a new GORM DB instance with the given context
func (c *Connection) WithContext(ctx context.Context) *gorm.DB {
	return c.DB.WithContext(ctx)
}

// AutoMigrate runs auto migration for given models
// This should only be used in development/testing environments
// Production should use migration files
func (c *Connection) AutoMigrate(models ...interface{}) error {
	if err := c.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	return nil
}

// GetDB returns the underlying GORM DB instance
func (c *Connection) GetDB() *gorm.DB {
	return c.DB
}

// IsConnected checks if the database connection is active
func (c *Connection) IsConnected() bool {
	if c.DB == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return c.HealthCheck(ctx) == nil
}
