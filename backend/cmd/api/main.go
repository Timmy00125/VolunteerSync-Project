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

	achievementHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/handlers"
	achievementRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/repositories"
	achievementServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/services"

	analyticsHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/analytics/handlers"
	analyticsServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/analytics/services"

	authHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/handlers"
	authRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/repositories"
	authServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/services"

	commHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/handlers"
	commRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/repositories"
	commServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/services"

	hoursHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/handlers"
	hoursRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/repositories"
	hoursServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/services"

	oppHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/handlers"
	oppRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/repositories"
	oppServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/services"

	orgHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/handlers"
	orgRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/repositories"
	orgServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/services"

	regHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/handlers"
	regRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/repositories"
	regServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/services"

	userHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/users/handlers"
	userServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/users/services"

	volunteerHandlers "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/handlers"
	volunteerRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/repositories"
	volunteerServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/services"

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

	// =============================================================================
	// Initialize Repositories
	// =============================================================================
	authRepo := authRepos.NewAuthRepository(dbConn.DB)
	orgRepo := orgRepos.NewOrganizationRepository(dbConn.DB)
	volunteerRepo := volunteerRepos.NewVolunteerRepository(dbConn.DB)
	oppRepo := oppRepos.NewOpportunityRepository(dbConn.DB)
	regRepo := regRepos.NewRegistrationRepository(dbConn.DB)
	hoursRepo := hoursRepos.NewHoursRepository(dbConn.DB)
	commRepo := commRepos.NewCommunicationsRepository(dbConn.DB)
	achievementRepo := achievementRepos.NewAchievementRepository(dbConn.DB)

	log.Info("All repositories initialized")

	// =============================================================================
	// Initialize Services (note: some services depend on other services)
	// =============================================================================

	// Auth service (no dependencies on other services)
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

	// User service (depends on authRepo)
	userService := userServices.NewUserService(
		authRepo,
		dbConn.DB,
		*log,
	)

	// Organization service (geocoding service is optional, passing nil)
	orgService := orgServices.NewOrganizationService(
		orgRepo,
		nil, // geocoding service - can be added later
		log,
	)

	// Volunteer service (geocoding service is optional, passing nil)
	volunteerService := volunteerServices.NewVolunteerService(
		volunteerRepo,
		nil, // geocoding service - can be added later
		log,
	)

	// Opportunity service (notification service is optional for now, passing nil)
	// Geocoding service is also optional
	oppService := oppServices.NewOpportunityService(
		oppRepo,
		nil, // geocoding service - can be added later
		nil, // notification service - can be added later
		log,
	)

	// =============================================================================
	// Create Adapters for Cross-Module Communication
	// =============================================================================

	// Opportunity adapter for registration service (capacity checks)
	oppAdapter := regServices.NewOpportunityServiceAdapter(oppService)

	// Registration service (depends on opportunity service via adapter)
	// Notification service is optional for now
	regService := regServices.NewRegistrationService(
		regRepo,
		oppAdapter, // opportunity service adapter - NOW WIRED
		nil,        // notification service - can be added later
		*log,
	)

	// Registration adapter for hours service (hour logging workflow)
	regAdapter := hoursServices.NewRegistrationServiceAdapter(regService)

	// Volunteer adapter for hours service (total hours increment)
	volunteerAdapter := hoursServices.NewVolunteerServiceAdapter(volunteerRepo)

	// Hours service (depends on registration and volunteer services via adapters)
	// Notification service is optional for now
	hoursService := hoursServices.NewHoursService(
		hoursRepo,
		regAdapter,       // registration service adapter - NOW WIRED
		volunteerAdapter, // volunteer service adapter - NOW WIRED
		nil,              // notification service - can be added later
		*log,
	)

	// Registration repository adapter for communications service (broadcast messages)
	regRepoAdapter := commServices.NewRegistrationRepositoryAdapter(regRepo)

	// Communications service (depends on registration repo via adapter for broadcast messages)
	commService := commServices.NewCommunicationsService(
		commRepo,
		regRepoAdapter, // registration repo adapter - NOW WIRED
		log,
	)

	// Analytics service (uses direct DB access)
	analyticsService := analyticsServices.NewAnalyticsService(
		dbConn.DB,
		*log,
	)

	// Communications service adapter for achievement service
	commServiceAdapter := achievementServices.NewCommunicationsServiceAdapter(commService)

	// Achievement service (depends on communications service for notifications)
	achievementService := achievementServices.NewAchievementService(
		achievementRepo,
		commServiceAdapter,
		log,
	)

	log.Info("All services initialized")

	// =============================================================================
	// Initialize Handlers
	// =============================================================================

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

	orgHandler, err := orgHandlers.NewOrganizationHandler(orgService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create organization handler")
	}

	volunteerHandler, err := volunteerHandlers.NewVolunteerHandler(volunteerService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create volunteer handler")
	}

	oppHandler, err := oppHandlers.NewOpportunityHandler(oppService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create opportunity handler")
	}

	regHandler, err := regHandlers.NewRegistrationHandler(regService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create registration handler")
	}

	hoursHandler, err := hoursHandlers.NewHoursHandler(hoursService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create hours handler")
	}

	commHandler, err := commHandlers.NewCommunicationsHandler(commService, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create communications handler")
	}

	analyticsHandler, err := analyticsHandlers.NewAnalyticsHandler(analyticsService, orgRepo, volunteerRepo, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create analytics handler")
	}

	achievementHandler, err := achievementHandlers.NewAchievementHandler(achievementService, orgRepo, log)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create achievement handler")
	}

	log.Info("All handlers initialized")

	// Register public routes (no authentication required)
	authGroup := v1.Group("/auth")
	authHandler.RegisterRoutes(authGroup)

	// Register protected routes (authentication required)
	// Create authenticated router group
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthMiddleware(jwtManager))
	authenticated.Use(middleware.ContextEnrichmentMiddleware()) // Convert user_id string to UUID
	authenticated.Use(middleware.RequireAnyRole())

	// User routes
	usersGroup := authenticated.Group("/users")
	userHandler.RegisterRoutes(usersGroup)

	// Organizations routes
	orgsGroup := authenticated.Group("/organizations")
	orgHandler.RegisterRoutes(orgsGroup)

	// Volunteers routes
	volunteersGroup := authenticated.Group("/volunteers")
	volunteerHandler.RegisterRoutes(volunteersGroup)

	// Opportunities routes (mixed: list/get are public with optional auth, create/update/delete require auth)
	publicOppsGroup := v1.Group("/opportunities")
	oppHandler.RegisterRoutes(publicOppsGroup)

	// Registrations routes
	regsGroup := authenticated.Group("/registrations")
	regHandler.RegisterRoutes(regsGroup)

	// Hours tracking routes
	hoursGroup := authenticated.Group("/hours")
	hoursHandler.RegisterRoutes(hoursGroup)

	// Communications routes (messages and notifications)
	commGroup := authenticated.Group("/communications")
	commHandler.RegisterRoutes(commGroup)

	// Achievements routes (mixed: list/get are public, create/award require auth)
	achievementsGroup := v1.Group("/achievements")
	achievementHandler.RegisterRoutes(achievementsGroup)

	// Analytics routes
	analyticsGroup := authenticated.Group("/analytics")
	analyticsHandler.RegisterRoutes(analyticsGroup)

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
