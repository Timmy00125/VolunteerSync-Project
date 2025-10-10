package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/repositories"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/cache"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/jwt"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// AuthService encapsulates authentication business logic, providing methods for
// user lifecycle workflows such as registration, login, token rotation, logout,
// and password reset. Handlers should depend on this interface to keep HTTP
// transport concerns isolated from domain logic.
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResponse, error)
	Login(ctx context.Context, input LoginInput) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	RequestPasswordReset(ctx context.Context, email string) (*PasswordResetChallenge, error)
	VerifySecurityAnswers(ctx context.Context, token string, answers []string) (*PasswordResetVerification, error)
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

// Config configures behaviour of the authentication service.
type Config struct {
	RefreshSessionPrefix        string
	PasswordResetSessionPrefix  string
	PasswordResetVerifyPrefix   string
	RefreshTokenTTL             time.Duration
	PasswordResetTTL            time.Duration
	PasswordResetVerifyTTL      time.Duration
	DefaultRole                 string
	MinSecurityAnswersToVerify  int
	MaxSecurityQuestionAttempts int
}

// DefaultConfig returns sane defaults aligned with platform requirements.
func DefaultConfig() Config {
	return Config{
		RefreshSessionPrefix:        "auth:refresh:",
		PasswordResetSessionPrefix:  "auth:pwd-reset:",
		PasswordResetVerifyPrefix:   "auth:pwd-verify:",
		RefreshTokenTTL:             7 * 24 * time.Hour,
		PasswordResetTTL:            15 * time.Minute,
		PasswordResetVerifyTTL:      15 * time.Minute,
		DefaultRole:                 "volunteer",
		MinSecurityAnswersToVerify:  2,
		MaxSecurityQuestionAttempts: 5,
	}
}

func applyConfigDefaults(cfg Config) Config {
	defaults := DefaultConfig()

	if cfg.RefreshSessionPrefix == "" {
		cfg.RefreshSessionPrefix = defaults.RefreshSessionPrefix
	}
	if cfg.PasswordResetSessionPrefix == "" {
		cfg.PasswordResetSessionPrefix = defaults.PasswordResetSessionPrefix
	}
	if cfg.PasswordResetVerifyPrefix == "" {
		cfg.PasswordResetVerifyPrefix = defaults.PasswordResetVerifyPrefix
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = defaults.RefreshTokenTTL
	}
	if cfg.PasswordResetTTL <= 0 {
		cfg.PasswordResetTTL = defaults.PasswordResetTTL
	}
	if cfg.PasswordResetVerifyTTL <= 0 {
		cfg.PasswordResetVerifyTTL = defaults.PasswordResetVerifyTTL
	}
	if cfg.DefaultRole == "" {
		cfg.DefaultRole = defaults.DefaultRole
	}
	if cfg.MinSecurityAnswersToVerify <= 0 {
		cfg.MinSecurityAnswersToVerify = defaults.MinSecurityAnswersToVerify
	}
	if cfg.MaxSecurityQuestionAttempts <= 0 {
		cfg.MaxSecurityQuestionAttempts = defaults.MaxSecurityQuestionAttempts
	}

	return cfg
}

// RegisterInput captures registration payload required to create a new user.
type RegisterInput struct {
	Email             string
	Password          string
	FirstName         string
	LastName          string
	Phone             *string
	UserType          string
	SecurityQuestions []SecurityQuestionInput
}

// SecurityQuestionInput captures a security question/answer pair.
type SecurityQuestionInput struct {
	Question string
	Answer   string
}

// LoginInput contains credentials for user login.
type LoginInput struct {
	Email    string
	Password string
}

// AuthUser represents sanitized user data returned by authentication flows.
type AuthUser struct {
	ID            uuid.UUID            `json:"id"`
	Email         string               `json:"email"`
	FirstName     string               `json:"first_name"`
	LastName      string               `json:"last_name"`
	Phone         *string              `json:"phone,omitempty"`
	AccountStatus models.AccountStatus `json:"account_status"`
	LastLoginAt   *time.Time           `json:"last_login_at,omitempty"`
	EmailVerified bool                 `json:"email_verified"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	UserType      string               `json:"user_type"`
}

// AuthResponse bundles user info with a freshly minted token pair.
type AuthResponse struct {
	User      *AuthUser      `json:"user"`
	TokenPair *jwt.TokenPair `json:"token_pair"`
}

// PasswordResetChallenge represents the first stage of password reset flow.
type PasswordResetChallenge struct {
	ResetToken        string   `json:"reset_token"`
	SecurityQuestions []string `json:"security_questions"`
}

// PasswordResetVerification represents the result of verifying security answers.
type PasswordResetVerification struct {
	VerifiedToken string `json:"verified_token"`
}

type authService struct {
	repo                 repositories.AuthRepository
	jwtManager           JWTManager
	log                  *logger.Logger
	refreshSessions      sessionStore
	resetSessions        sessionStore
	verificationSessions sessionStore
	config               Config
	now                  func() time.Time
}

type sessionStore interface {
	SetSession(ctx context.Context, sessionID string, data interface{}) error
	GetSession(ctx context.Context, sessionID string, dest interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// JWTManager defines the subset of token manager behaviour required by AuthService.
type JWTManager interface {
	GenerateTokenPair(userID, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
	RefreshTokenPair(refreshToken, userRole string) (*jwt.TokenPair, string, error)
	GetTokenID(token string) (string, error)
}

type refreshSessionData struct {
	UserID    string    `json:"user_id"`
	UserRole  string    `json:"user_role"`
	TokenID   string    `json:"token_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type passwordResetSessionData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Attempts  int       `json:"attempts"`
}

type passwordResetVerificationData struct {
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	VerifiedAt time.Time `json:"verified_at"`
}

// NewAuthService constructs an AuthService with required dependencies.
func NewAuthService(
	repo repositories.AuthRepository,
	jwtManager JWTManager,
	cacheClient *cache.Client,
	cfg Config,
	log *logger.Logger,
) (AuthService, error) {
	if cacheClient == nil {
		return nil, fmt.Errorf("auth service requires cache client")
	}

	cfg = applyConfigDefaults(cfg)

	refreshStore := cache.NewSessionStorage(cacheClient, cfg.RefreshSessionPrefix, cfg.RefreshTokenTTL)
	resetStore := cache.NewSessionStorage(cacheClient, cfg.PasswordResetSessionPrefix, cfg.PasswordResetTTL)
	verifyStore := cache.NewSessionStorage(cacheClient, cfg.PasswordResetVerifyPrefix, cfg.PasswordResetVerifyTTL)

	return NewAuthServiceWithStores(repo, jwtManager, refreshStore, resetStore, verifyStore, cfg, log)
}

// NewAuthServiceWithStores allows injecting custom session stores (useful for testing).
func NewAuthServiceWithStores(
	repo repositories.AuthRepository,
	jwtManager JWTManager,
	refreshStore sessionStore,
	resetStore sessionStore,
	verificationStore sessionStore,
	cfg Config,
	log *logger.Logger,
) (AuthService, error) {
	if repo == nil {
		return nil, fmt.Errorf("auth service requires repository")
	}
	if jwtManager == nil {
		return nil, fmt.Errorf("auth service requires jwt manager")
	}
	if refreshStore == nil || resetStore == nil || verificationStore == nil {
		return nil, fmt.Errorf("auth service requires session stores")
	}
	if log == nil {
		log = logger.Get()
	}

	cfg = applyConfigDefaults(cfg)

	service := &authService{
		repo:                 repo,
		jwtManager:           jwtManager,
		log:                  log,
		refreshSessions:      refreshStore,
		resetSessions:        resetStore,
		verificationSessions: verificationStore,
		config:               cfg,
		now:                  time.Now,
	}

	return service, nil
}

// Register creates a new user, hashes credentials, persists the account, and issues tokens.
func (s *authService) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	validationErrs := s.validateRegisterInput(input)
	if len(validationErrs) > 0 {
		return nil, apperrors.NewValidationError("invalid registration payload", validationErrs)
	}

	email := normalizeEmail(input.Email)

	user := &models.User{
		Email:         email,
		FirstName:     strings.TrimSpace(input.FirstName),
		LastName:      strings.TrimSpace(input.LastName),
		AccountStatus: models.AccountStatusActive,
		EmailVerified: false,
	}

	if input.Phone != nil {
		trimmed := strings.TrimSpace(*input.Phone)
		if trimmed != "" {
			user.Phone = &trimmed
		}
	}

	if err := user.SetPassword(input.Password); err != nil {
		return nil, mapPasswordError(err)
	}

	q1 := strings.TrimSpace(input.SecurityQuestions[0].Question)
	a1 := strings.TrimSpace(input.SecurityQuestions[0].Answer)
	q2 := strings.TrimSpace(input.SecurityQuestions[1].Question)
	a2 := strings.TrimSpace(input.SecurityQuestions[1].Answer)
	q3 := strings.TrimSpace(input.SecurityQuestions[2].Question)
	a3 := strings.TrimSpace(input.SecurityQuestions[2].Answer)

	if err := user.SetSecurityAnswers(q1, a1, q2, a2, q3, a3); err != nil {
		return nil, apperrors.NewInternalServerError("failed to store security answers").WithError(err)
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return nil, apperrors.NewConflictError("email already registered")
		}
		return nil, apperrors.NewInternalServerError("failed to create user").WithError(err)
	}

	role := input.UserType
	if role == "" {
		role = s.config.DefaultRole
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID.String(), role)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to generate tokens").WithError(err)
	}

	if err := s.storeRefreshSession(ctx, tokenPair.RefreshToken, user.ID, role); err != nil {
		return nil, apperrors.NewInternalServerError("failed to persist refresh session").WithError(err)
	}

	s.log.WithContext(ctx).LogAuthentication(user.ID.String(), "register", true)

	return &AuthResponse{
		User:      buildAuthUserDTO(user, role),
		TokenPair: tokenPair,
	}, nil
}

// Login authenticates user credentials, updates account metadata, and issues a token pair.
func (s *authService) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	validationErrs := s.validateLoginInput(input)
	if len(validationErrs) > 0 {
		return nil, apperrors.NewValidationError("invalid login payload", validationErrs)
	}

	email := normalizeEmail(input.Email)
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, apperrors.NewUnauthorizedError("invalid credentials")
		}
		return nil, apperrors.NewInternalServerError("failed to retrieve user").WithError(err)
	}

	if err := user.VerifyPassword(input.Password); err != nil {
		if errors.Is(err, models.ErrInvalidPassword) {
			return nil, apperrors.NewUnauthorizedError("invalid credentials")
		}
		return nil, apperrors.NewInternalServerError("failed to verify password").WithError(err)
	}

	if user.AccountStatus == models.AccountStatusSuspended {
		return nil, apperrors.NewForbiddenError("account is suspended")
	}

	role := s.config.DefaultRole

	if user.AccountStatus == models.AccountStatusInactive {
		if err := s.repo.UpdateAccountStatus(ctx, user.ID, models.AccountStatusActive); err != nil {
			return nil, apperrors.NewInternalServerError("failed to update account status").WithError(err)
		}
		user.AccountStatus = models.AccountStatusActive
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.log.WithContext(ctx).Errorf("failed to update last login for user %s: %v", user.ID.String(), err)
	} else {
		now := s.now()
		user.LastLoginAt = &now
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID.String(), role)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to generate tokens").WithError(err)
	}

	if err := s.storeRefreshSession(ctx, tokenPair.RefreshToken, user.ID, role); err != nil {
		return nil, apperrors.NewInternalServerError("failed to persist refresh session").WithError(err)
	}

	s.log.WithContext(ctx).LogAuthentication(user.ID.String(), "login", true)

	return &AuthResponse{
		User:      buildAuthUserDTO(user, role),
		TokenPair: tokenPair,
	}, nil
}

// RefreshToken performs refresh token rotation, invalidating the old token and returning a new pair.
// This also extends the session TTL (sliding window) to keep active users logged in.
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, apperrors.NewBadRequestError("refresh token is required")
	}

	tokenID, err := s.jwtManager.GetTokenID(refreshToken)
	if err != nil {
		return nil, apperrors.NewUnauthorizedError("invalid refresh token")
	}

	session, err := s.getRefreshSession(ctx, tokenID)
	if err != nil {
		s.log.WithContext(ctx).Warnf("failed to retrieve refresh session %s: %v", tokenID, err)
		return nil, apperrors.NewUnauthorizedError("refresh token has been revoked")
	}

	// Validate that the refresh token hasn't been tampered with
	tokenPair, oldTokenID, err := s.jwtManager.RefreshTokenPair(refreshToken, session.UserRole)
	if err != nil {
		s.log.WithContext(ctx).Warnf("failed to generate new token pair: %v", err)
		return nil, apperrors.NewUnauthorizedError("failed to refresh token")
	}

	if oldTokenID != session.TokenID {
		s.log.WithContext(ctx).Warnf("token ID mismatch: expected %s, got %s", session.TokenID, oldTokenID)
		return nil, apperrors.NewUnauthorizedError("refresh token mismatch")
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return nil, apperrors.NewInternalServerError("invalid session data").WithError(err)
	}

	// Delete old session FIRST to prevent reuse attack if storage of new session fails
	if err := s.refreshSessions.DeleteSession(ctx, session.TokenID); err != nil {
		s.log.WithContext(ctx).Errorf("failed to delete old refresh session %s: %v", session.TokenID, err)
		// Continue anyway - better to have duplicate session than fail refresh
	}

	// Store new session with full TTL (implements sliding window for active users)
	if err := s.storeRefreshSession(ctx, tokenPair.RefreshToken, userID, session.UserRole); err != nil {
		s.log.WithContext(ctx).Errorf("failed to persist new refresh session: %v", err)
		return nil, apperrors.NewInternalServerError("failed to persist refresh session").WithError(err)
	}

	s.log.WithContext(ctx).
		WithField("user_id", session.UserID).
		WithField("old_token_id", oldTokenID).
		Info("token refresh successful - session extended")

	return tokenPair, nil
}

// Logout revokes the provided refresh token, making it unusable for future refresh attempts.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return apperrors.NewBadRequestError("refresh token is required")
	}

	tokenID, err := s.jwtManager.GetTokenID(refreshToken)
	if err != nil {
		return apperrors.NewUnauthorizedError("invalid refresh token")
	}

	session, err := s.getRefreshSession(ctx, tokenID)
	if err != nil {
		return apperrors.NewUnauthorizedError("refresh token already invalidated")
	}

	if err := s.refreshSessions.DeleteSession(ctx, tokenID); err != nil {
		return apperrors.NewUnauthorizedError("refresh token already invalidated")
	}

	s.log.WithContext(ctx).LogAuthentication(session.UserID, "logout", true)
	return nil
}

// RequestPasswordReset begins the password reset flow by returning security questions.
func (s *authService) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetChallenge, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, apperrors.NewBadRequestError("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, apperrors.NewValidationError("invalid email", map[string]interface{}{"email": "invalid email format"})
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, apperrors.NewNotFoundError("user")
		}
		return nil, apperrors.NewInternalServerError("failed to retrieve user").WithError(err)
	}

	resetToken, err := generateSecureToken(32)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to create reset token").WithError(err)
	}

	session := &passwordResetSessionData{
		UserID:    user.ID.String(),
		Email:     user.Email,
		CreatedAt: s.now(),
		Attempts:  0,
	}

	if err := s.resetSessions.SetSession(ctx, resetToken, session); err != nil {
		return nil, apperrors.NewInternalServerError("failed to persist reset session").WithError(err)
	}

	questions := user.GetSecurityQuestions()

	return &PasswordResetChallenge{
		ResetToken:        resetToken,
		SecurityQuestions: questions,
	}, nil
}

// VerifySecurityAnswers validates security question answers and issues a verification token for password reset.
func (s *authService) VerifySecurityAnswers(ctx context.Context, token string, answers []string) (*PasswordResetVerification, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, apperrors.NewBadRequestError("reset token is required")
	}
	if len(answers) != 3 {
		return nil, apperrors.NewValidationError("invalid answers", map[string]interface{}{"answers": "exactly 3 answers are required"})
	}

	session := &passwordResetSessionData{}
	if err := s.resetSessions.GetSession(ctx, token, session); err != nil {
		return nil, apperrors.NewBadRequestError("invalid or expired reset token")
	}

	if session.Attempts >= s.config.MaxSecurityQuestionAttempts {
		_ = s.resetSessions.DeleteSession(ctx, token)
		return nil, apperrors.NewForbiddenError("maximum verification attempts exceeded")
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		_ = s.resetSessions.DeleteSession(ctx, token)
		return nil, apperrors.NewInternalServerError("invalid session data").WithError(err)
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		_ = s.resetSessions.DeleteSession(ctx, token)
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, apperrors.NewNotFoundError("user")
		}
		return nil, apperrors.NewInternalServerError("failed to retrieve user").WithError(err)
	}

	answersMap := map[int]string{
		1: strings.TrimSpace(answers[0]),
		2: strings.TrimSpace(answers[1]),
		3: strings.TrimSpace(answers[2]),
	}

	correctCount, verifyErr := user.VerifySecurityAnswers(answersMap)
	if verifyErr != nil && !errors.Is(verifyErr, models.ErrInvalidSecurityAnswer) {
		return nil, apperrors.NewInternalServerError("failed to verify security answers").WithError(verifyErr)
	}

	session.Attempts++
	if err := s.resetSessions.SetSession(ctx, token, session); err != nil {
		return nil, apperrors.NewInternalServerError("failed to update reset session").WithError(err)
	}

	if correctCount < s.config.MinSecurityAnswersToVerify {
		return nil, apperrors.NewBadRequestError("insufficient correct answers")
	}

	verifiedToken, err := generateSecureToken(32)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to generate verification token").WithError(err)
	}

	verificationSession := &passwordResetVerificationData{
		UserID:     user.ID.String(),
		Email:      user.Email,
		VerifiedAt: s.now(),
	}

	if err := s.verificationSessions.SetSession(ctx, verifiedToken, verificationSession); err != nil {
		return nil, apperrors.NewInternalServerError("failed to persist verification token").WithError(err)
	}

	_ = s.resetSessions.DeleteSession(ctx, token)

	return &PasswordResetVerification{VerifiedToken: verifiedToken}, nil
}

// ConfirmPasswordReset sets a new password for the user after verification.
func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return apperrors.NewBadRequestError("verified token is required")
	}

	if err := models.ValidatePasswordStrength(newPassword); err != nil {
		return mapPasswordError(err)
	}

	verification := &passwordResetVerificationData{}
	if err := s.verificationSessions.GetSession(ctx, token, verification); err != nil {
		return apperrors.NewBadRequestError("invalid or expired verification token")
	}

	userID, err := uuid.Parse(verification.UserID)
	if err != nil {
		_ = s.verificationSessions.DeleteSession(ctx, token)
		return apperrors.NewInternalServerError("invalid verification data").WithError(err)
	}

	tempUser := &models.User{}
	if err := tempUser.SetPassword(newPassword); err != nil {
		return mapPasswordError(err)
	}

	if err := s.repo.UpdatePassword(ctx, userID, tempUser.PasswordHash); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return apperrors.NewNotFoundError("user")
		}
		return apperrors.NewInternalServerError("failed to update password").WithError(err)
	}

	if err := s.verificationSessions.DeleteSession(ctx, token); err != nil {
		s.log.WithContext(ctx).Warnf("failed to delete verification session for user %s: %v", userID, err)
	}

	s.log.WithContext(ctx).LogAuthentication(userID.String(), "password_reset", true)

	return nil
}

func (s *authService) validateRegisterInput(input RegisterInput) map[string]interface{} {
	errorsMap := make(map[string]interface{})

	if strings.TrimSpace(input.Email) == "" {
		errorsMap["email"] = "email is required"
	} else if _, err := mail.ParseAddress(strings.TrimSpace(input.Email)); err != nil {
		errorsMap["email"] = "invalid email format"
	}

	if strings.TrimSpace(input.FirstName) == "" {
		errorsMap["first_name"] = "first_name is required"
	}

	if strings.TrimSpace(input.LastName) == "" {
		errorsMap["last_name"] = "last_name is required"
	}

	if strings.TrimSpace(input.Password) == "" {
		errorsMap["password"] = "password is required"
	} else if err := models.ValidatePasswordStrength(input.Password); err != nil {
		errorsMap["password"] = err.Error()
	}

	if strings.TrimSpace(input.UserType) == "" {
		errorsMap["user_type"] = "user_type is required"
	} else if !isSupportedUserType(strings.TrimSpace(input.UserType)) {
		errorsMap["user_type"] = "unsupported user_type"
	}

	if len(input.SecurityQuestions) != 3 {
		errorsMap["security_questions"] = "exactly 3 security questions are required"
	} else {
		for idx, q := range input.SecurityQuestions {
			if strings.TrimSpace(q.Question) == "" || strings.TrimSpace(q.Answer) == "" {
				errorsMap[fmt.Sprintf("security_questions[%d]", idx)] = "question and answer are required"
			}
		}
	}

	return errorsMap
}

func (s *authService) validateLoginInput(input LoginInput) map[string]interface{} {
	errorsMap := make(map[string]interface{})

	if strings.TrimSpace(input.Email) == "" {
		errorsMap["email"] = "email is required"
	}
	if strings.TrimSpace(input.Password) == "" {
		errorsMap["password"] = "password is required"
	}

	return errorsMap
}

func (s *authService) storeRefreshSession(ctx context.Context, refreshToken string, userID uuid.UUID, role string) error {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	tokenID := claims.ID
	if tokenID == "" {
		return fmt.Errorf("refresh token missing ID")
	}

	data := &refreshSessionData{
		UserID:    userID.String(),
		UserRole:  role,
		TokenID:   tokenID,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if err := s.refreshSessions.SetSession(ctx, tokenID, data); err != nil {
		s.log.WithContext(ctx).
			WithField("user_id", userID.String()).
			WithField("token_id", tokenID).
			Errorf("failed to store refresh session: %v", err)
		return err
	}

	s.log.WithContext(ctx).
		WithField("user_id", userID.String()).
		WithField("token_id", tokenID).
		WithField("expires_at", claims.ExpiresAt.Time.Format(time.RFC3339)).
		Debug("refresh session stored successfully")

	return nil
}

func (s *authService) getRefreshSession(ctx context.Context, tokenID string) (*refreshSessionData, error) {
	session := &refreshSessionData{}
	if err := s.refreshSessions.GetSession(ctx, tokenID, session); err != nil {
		return nil, err
	}
	return session, nil
}

func buildAuthUserDTO(user *models.User, userType string) *AuthUser {
	dto := &AuthUser{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		AccountStatus: user.AccountStatus,
		LastLoginAt:   user.LastLoginAt,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		UserType:      userType,
	}
	return dto
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func generateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func isSupportedUserType(userType string) bool {
	switch strings.ToLower(userType) {
	case "volunteer", "organization_admin":
		return true
	default:
		return false
	}
}

func mapPasswordError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, models.ErrPasswordTooWeak) {
		return apperrors.NewValidationError("password does not meet complexity requirements", map[string]interface{}{"password": err.Error()})
	}

	return apperrors.NewInternalServerError("password processing failed").WithError(err)
}
