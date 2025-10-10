package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// AuthHandlerConfig controls rate limiting behaviour for auth endpoints.
type AuthHandlerConfig struct {
	RegistrationLimit  int64
	RegistrationWindow time.Duration
	LoginLimit         int64
	LoginWindow        time.Duration
	RefreshLimit       int64
	RefreshWindow      time.Duration
}

// DefaultAuthHandlerConfig returns opinionated defaults aligned with platform requirements.
func DefaultAuthHandlerConfig() AuthHandlerConfig {
	return AuthHandlerConfig{
		RegistrationLimit:  5,
		RegistrationWindow: 15 * time.Minute,
		LoginLimit:         5,
		LoginWindow:        15 * time.Minute,
		RefreshLimit:       20,  // More generous for refresh tokens
		RefreshWindow:      15 * time.Minute,
	}
}

// AuthHandler exposes HTTP handlers for authentication flows.
type AuthHandler struct {
	service     services.AuthService
	rateLimiter RateLimiter
	log         *logger.Logger
	config      AuthHandlerConfig
}

// NewAuthHandler constructs an AuthHandler with required dependencies.
func NewAuthHandler(service services.AuthService, rateLimiter RateLimiter, log *logger.Logger, cfg AuthHandlerConfig) (*AuthHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("auth handler requires auth service")
	}

	if cfg.RegistrationLimit == 0 {
		cfg.RegistrationLimit = DefaultAuthHandlerConfig().RegistrationLimit
	}
	if cfg.RegistrationWindow == 0 {
		cfg.RegistrationWindow = DefaultAuthHandlerConfig().RegistrationWindow
	}
	if cfg.LoginLimit == 0 {
		cfg.LoginLimit = DefaultAuthHandlerConfig().LoginLimit
	}
	if cfg.LoginWindow == 0 {
		cfg.LoginWindow = DefaultAuthHandlerConfig().LoginWindow
	}
	if cfg.RefreshLimit == 0 {
		cfg.RefreshLimit = DefaultAuthHandlerConfig().RefreshLimit
	}
	if cfg.RefreshWindow == 0 {
		cfg.RefreshWindow = DefaultAuthHandlerConfig().RefreshWindow
	}

	if log == nil {
		log = logger.Get()
	}

	if rateLimiter == nil {
		rateLimiter = noopRateLimiter{}
	}

	return &AuthHandler{
		service:     service,
		rateLimiter: rateLimiter,
		log:         log,
		config:      cfg,
	}, nil
}

// RegisterRoutes wires authentication routes under the provided router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.RefreshToken)
	rg.POST("/logout", h.Logout)
	rg.POST("/password-reset/request", h.RequestPasswordReset)
	rg.POST("/password-reset/verify", h.VerifySecurityAnswers)
	rg.POST("/password-reset/confirm", h.ConfirmPasswordReset)
}

type registerRequest struct {
	Email             string                   `json:"email"`
	Password          string                   `json:"password"`
	FirstName         string                   `json:"first_name"`
	LastName          string                   `json:"last_name"`
	Phone             *string                  `json:"phone"`
	UserType          string                   `json:"user_type"`
	SecurityQuestions []securityQuestionIntent `json:"security_questions"`
}

type securityQuestionIntent struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
}

type passwordResetVerifyRequest struct {
	ResetToken string   `json:"reset_token"`
	Answers    []string `json:"answers"`
}

type passwordResetConfirmRequest struct {
	VerifiedToken string `json:"verified_token"`
	NewPassword   string `json:"new_password"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	if !h.allowRequest(c, "registration", h.config.RegistrationLimit, h.config.RegistrationWindow) {
		return
	}

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	var phonePtr *string
	if req.Phone != nil {
		trimmed := strings.TrimSpace(*req.Phone)
		if trimmed != "" {
			phonePtr = &trimmed
		}
	}

	questions := make([]services.SecurityQuestionInput, 0, len(req.SecurityQuestions))
	for _, q := range req.SecurityQuestions {
		questions = append(questions, services.SecurityQuestionInput{
			Question: q.Question,
			Answer:   q.Answer,
		})
	}

	input := services.RegisterInput{
		Email:             req.Email,
		Password:          req.Password,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Phone:             phonePtr,
		UserType:          req.UserType,
		SecurityQuestions: questions,
	}

	resp, err := h.service.Register(ctx, input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":          resp.User,
		"access_token":  resp.TokenPair.AccessToken,
		"refresh_token": resp.TokenPair.RefreshToken,
		"expires_in":    resp.TokenPair.ExpiresIn,
	})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	if !h.allowRequest(c, "login", h.config.LoginLimit, h.config.LoginWindow) {
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	resp, err := h.service.Login(ctx, services.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          resp.User,
		"access_token":  resp.TokenPair.AccessToken,
		"refresh_token": resp.TokenPair.RefreshToken,
		"expires_in":    resp.TokenPair.ExpiresIn,
	})
}

// RefreshToken handles POST /auth/refresh.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Apply rate limiting to prevent token refresh abuse
	if !h.allowRequest(c, "refresh", h.config.RefreshLimit, h.config.RefreshWindow) {
		return
	}

	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	tokens, err := h.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authorization header required"))
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid authorization header"))
		return
	}

	refreshToken := strings.TrimSpace(parts[1])
	if refreshToken == "" {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid authorization header"))
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Logout(ctx, refreshToken); err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RequestPasswordReset handles POST /auth/password-reset/request.
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req passwordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	challenge, err := h.service.RequestPasswordReset(ctx, req.Email)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reset_token":        challenge.ResetToken,
		"security_questions": challenge.SecurityQuestions,
	})
}

// VerifySecurityAnswers handles POST /auth/password-reset/verify.
func (h *AuthHandler) VerifySecurityAnswers(c *gin.Context) {
	var req passwordResetVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	verification, err := h.service.VerifySecurityAnswers(ctx, req.ResetToken, req.Answers)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified_token": verification.VerifiedToken,
	})
}

// ConfirmPasswordReset handles POST /auth/password-reset/confirm.
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req passwordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	if err := h.service.ConfirmPasswordReset(ctx, req.VerifiedToken, req.NewPassword); err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (h *AuthHandler) allowRequest(c *gin.Context, action string, limit int64, window time.Duration) bool {
	if limit <= 0 || window <= 0 || h.rateLimiter == nil {
		return true
	}

	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}

	key := fmt.Sprintf("auth:%s:%s", action, ip)

	allowed, retryAfter, err := h.rateLimiter.Allow(c.Request.Context(), key, limit, window)
	if err != nil {
		h.log.WithContext(c.Request.Context()).Warnf("rate limiter error for %s: %v", key, err)
		return true
	}

	if allowed {
		return true
	}

	seconds := int64(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}

	c.Header("Retry-After", strconv.FormatInt(seconds, 10))
	h.respondWithError(c, apperrors.NewRateLimitError(fmt.Sprintf("too many %s attempts, please try again later", action)))
	return false
}

func (h *AuthHandler) respondWithError(c *gin.Context, err error) {
	apperrors.AbortWithError(c, err)
}
