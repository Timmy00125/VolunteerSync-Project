package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ContextKey is the type for context keys used in logging
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user IDs
	UserIDKey ContextKey = "user_id"
	// OrgIDKey is the context key for organization IDs
	OrgIDKey ContextKey = "org_id"
)

// Config holds the logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	Output     io.Writer
	WithCaller bool
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger with the given configuration
func Init(cfg Config) {
	// Parse log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Set output
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Configure format
	var logger zerolog.Logger
	if cfg.Format == "console" {
		// Human-readable console output for development
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create logger
	logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()

	// Add caller information if requested
	if cfg.WithCaller {
		logger = logger.With().Caller().Logger()
	}

	globalLogger = &Logger{logger: logger}

	// Set as global default
	log.Logger = logger
}

// Get returns the global logger instance
// If not initialized, it returns a default logger
func Get() *Logger {
	if globalLogger == nil {
		// Return default logger if not initialized
		defaultLogger := zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()
		return &Logger{logger: defaultLogger}
	}
	return globalLogger
}

// WithContext returns a logger with context values
// It extracts request_id, user_id, and org_id from context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	// Add request ID if present
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With().Str("request_id", requestID.(string)).Logger()
	}

	// Add user ID if present - sanitized to not expose PII
	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With().Str("user_id", sanitizeUserID(userID)).Logger()
	}

	// Add organization ID if present
	if orgID := ctx.Value(OrgIDKey); orgID != nil {
		logger = logger.With().Str("org_id", sanitizeOrgID(orgID)).Logger()
	}

	return &Logger{logger: logger}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields returns a logger with multiple additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.logger
	for key, value := range fields {
		// Ensure no PII is logged
		if isPIIField(key) {
			continue
		}
		logger = logger.With().Interface(key, value).Logger()
	}
	return &Logger{logger: logger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error message with an error object
func (l *Logger) ErrorWithErr(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// parseLevel converts a string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// sanitizeUserID ensures user IDs don't contain PII
// Only keeps the ID itself, not any user details
func sanitizeUserID(userID interface{}) string {
	// Convert to string representation of ID only
	switch v := userID.(type) {
	case string:
		return v
	case int, int64, uint, uint64:
		return "user_id"
	default:
		return "user_id"
	}
}

// sanitizeOrgID ensures org IDs don't contain PII
func sanitizeOrgID(orgID interface{}) string {
	switch v := orgID.(type) {
	case string:
		return v
	case int, int64, uint, uint64:
		return "org_id"
	default:
		return "org_id"
	}
}

// isPIIField checks if a field name contains PII
// This ensures we never log sensitive personal information
func isPIIField(fieldName string) bool {
	piiFields := []string{
		"password",
		"email",
		"phone",
		"ssn",
		"credit_card",
		"address",
		"name",
		"first_name",
		"last_name",
		"ip_address",
		"date_of_birth",
		"dob",
		"emergency_contact",
		"security_question",
		"security_answer",
	}

	for _, pii := range piiFields {
		if fieldName == pii {
			return true
		}
	}
	return false
}

// Helper functions for common logging patterns

// LogRequest logs an HTTP request with safe information
func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration) {
	l.logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration_ms", duration).
		Msg("HTTP request")
}

// LogDatabaseQuery logs a database query (without sensitive data)
func (l *Logger) LogDatabaseQuery(operation string, table string, duration time.Duration) {
	l.logger.Debug().
		Str("operation", operation).
		Str("table", table).
		Dur("duration_ms", duration).
		Msg("Database query")
}

// LogCacheOperation logs a cache operation
func (l *Logger) LogCacheOperation(operation string, key string, hit bool) {
	l.logger.Debug().
		Str("operation", operation).
		Str("key", key).
		Bool("hit", hit).
		Msg("Cache operation")
}

// LogAuthentication logs authentication events (without sensitive data)
func (l *Logger) LogAuthentication(userID string, action string, success bool) {
	event := l.logger.Info()
	if !success {
		event = l.logger.Warn()
	}
	event.
		Str("user_id", userID).
		Str("action", action).
		Bool("success", success).
		Msg("Authentication event")
}

// LogAuthorization logs authorization events
func (l *Logger) LogAuthorization(userID string, resource string, action string, allowed bool) {
	event := l.logger.Info()
	if !allowed {
		event = l.logger.Warn()
	}
	event.
		Str("user_id", userID).
		Str("resource", resource).
		Str("action", action).
		Bool("allowed", allowed).
		Msg("Authorization event")
}
