package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"

	authHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/handlers"
	authRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/repositories"
	authServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/services"

	userHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/users/handlers"
	userServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/users/services"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/cache"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/database"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/jwt"
	appLogger "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// getEnv retrieves an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize logger
	logLevel := getEnv("LOG_LEVEL", "info")
	logFormat := getEnv("LOG_FORMAT", "json")
	appLogger.Init(appLogger.Config{
		Level:      logLevel,
		Format:     logFormat,
		WithCaller: true,
	})
	log := appLogger.Get()

	log.Info("Starting VolunteerSync API server...")

	// Initialize database connection
	dbConfig := &database.Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "volunteersync"),
		Password:        getEnv("DB_PASSWORD", "volunteersync"),
		DBName:          getEnv("DB_NAME", "volunteersync"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        logger.Info,
	}

	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to connect to database")
	}
	log.Info("Database connection established")

	// Initialize Redis connection
	redisConfig := &cache.Config{
		Host:            getEnv("REDIS_HOST", "localhost"),
		Port:            getEnv("REDIS_PORT", "6379"),
		Password:        getEnv("REDIS_PASSWORD", ""),
		DB:              0,
		MaxRetries:      3,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}

	redisClient, err := cache.NewClient(redisConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to connect to Redis")
	}
	log.Info("Redis connection established")

	// Initialize JWT manager
	jwtConfig := &jwt.Config{
		AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-access-secret-change-in-production"),
		RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             getEnv("JWT_ISSUER", "volunteersync"),
	}
	jwtManager := jwt.NewManager(jwtConfig)
	log.Info("JWT manager initialized")

	// Set Gin mode
	ginMode := getEnv("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	// Create Gin router
	router := gin.New()

	// Setup middleware chain: logging → recovery → CORS → rate limiting → auth → RBAC
	// 1. Logging middleware (logs all requests)
	router.Use(middleware.LoggingMiddleware())

	// 2. Recovery middleware (catch panics)
	router.Use(middleware.RecoveryMiddleware())

	// 3. CORS middleware
	corsConfig := middleware.DefaultCORSConfig()
	// Allow additional origins from environment
	if additionalOrigins := getEnv("CORS_ALLOWED_ORIGINS", ""); additionalOrigins != "" {
		corsConfig.AllowedOrigins = append(corsConfig.AllowedOrigins, additionalOrigins)
	}
	router.Use(middleware.CORSMiddleware(corsConfig))

	// 4. General rate limiting (100 requests per minute)
	router.Use(middleware.RateLimitMiddleware(redisClient, middleware.DefaultRateLimitConfig()))

	log.Info("Middleware chain configured")

	// Health check endpoint (no auth required)
	router.GET("/health", healthCheckHandler(dbConn, redisClient))

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Initialize repositories
	authRepo := authRepos.NewAuthRepository(dbConn.DB)

	// Initialize services
	authServiceConfig := authServices.DefaultConfig()
	authService, err := authServices.NewAuthService(
		authRepo,
		jwtManager,
		redisClient,
		authServiceConfig,
		log,
	)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create auth service")
	}

	userService := userServices.NewUserService(
		authRepo,
		dbConn.DB,
		*log,
	)

	// Initialize handlers
	authHandler, err := authHandlers.NewAuthHandler(
		authService,
		&redisRateLimiterAdapter{client: redisClient},
		log,
		authHandlers.DefaultAuthHandlerConfig(),
	)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create auth handler")
	}

	userHandler, err := userHandlers.NewUserHandler(userService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create user handler")
	}

	log.Info("All handlers initialized")

	// Register public routes (no authentication required)
	authGroup := v1.Group("/auth")
	authHandler.RegisterRoutes(authGroup)

	// Register protected routes (authentication required)
	// Create authenticated router group
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthMiddleware(jwtManager))
	authenticated.Use(middleware.RequireAnyRole())

	// User routes
	usersGroup := authenticated.Group("/users")
	userHandler.RegisterRoutes(usersGroup)

	// TODO: Register additional module routes as they are completed
	// - Organizations
	// - Volunteers
	// - Opportunities
	// - Registrations
	// - Hours tracking
	// - Communications
	// - Achievements
	// - Analytics

	log.Info("All routes registered")

	// Start HTTP server
	port := getEnv("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)

	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.WithField("address", addr).Info("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithField("error", err.Error()).Fatal("Failed to start HTTP server")
		}
	}()

	log.WithField("port", port).Info("Server started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithField("error", err.Error()).Fatal("Server forced to shutdown")
	}

	// Close database connection
	sqlDB, err := dbConn.DB.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.WithField("error", err.Error()).Error("Failed to close database connection")
		} else {
			log.Info("Database connection closed")
		}
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.WithField("error", err.Error()).Error("Failed to close Redis connection")
	} else {
		log.Info("Redis connection closed")
	}

	log.Info("Server shutdown complete")
}

// healthCheckHandler returns a handler for the health check endpoint
// It checks the status of the database and Redis connections
func healthCheckHandler(db *database.Connection, redis *cache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := "healthy"
		statusCode := http.StatusOK
		checks := make(map[string]string)

		// Check database connection
		sqlDB, err := db.DB.DB()
		if err != nil {
			checks["database"] = "error: failed to get underlying DB"
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else if err := sqlDB.PingContext(ctx); err != nil {
			checks["database"] = "error: " + err.Error()
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else {
			checks["database"] = "ok"
		}

		// Check Redis connection
		redisInternal := redis.GetClient()
		if err := redisInternal.Ping(ctx).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else {
			checks["redis"] = "ok"
		}

		c.JSON(statusCode, gin.H{
			"status":    status,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"checks":    checks,
		})
	}
}

// redisRateLimiterAdapter adapts cache.Client to implement auth handlers' RateLimiter interface
type redisRateLimiterAdapter struct {
	client *cache.Client
}

func (r *redisRateLimiterAdapter) Allow(ctx context.Context, identifier string, limit int64, window time.Duration) (bool, time.Duration, error) {
	key := fmt.Sprintf("rate_limit:auth:%s", identifier)

	count, err := r.client.Increment(ctx, key)
	if err != nil {
		// If Redis fails, allow the request (fail open)
		return true, 0, nil
	}

	// Set expiration on first request
	if count == 1 {
		if err := r.client.Expire(ctx, key, window); err != nil {
			// Log error but don't block request
			return true, 0, nil
		}
	}

	if count > limit {
		// Get remaining TTL for retry after duration
		ttl, err := r.client.TTL(ctx, key)
		if err != nil {
			ttl = window // Fallback to window duration
		}
		return false, ttl, nil
	}

	return true, 0, nil
}
