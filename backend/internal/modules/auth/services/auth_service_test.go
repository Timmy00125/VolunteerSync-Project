package services

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/repositories"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/jwt"
)

type mockAuthRepository struct {
	mock.Mock
}

func (m *mockAuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockAuthRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *mockAuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *mockAuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *mockAuthRepository) GetSecurityQuestions(ctx context.Context, email string) ([]string, error) {
	args := m.Called(ctx, email)
	questions, _ := args.Get(0).([]string)
	return questions, args.Error(1)
}

func (m *mockAuthRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockAuthRepository) UpdateAccountStatus(ctx context.Context, userID uuid.UUID, status models.AccountStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

type mockJWTManager struct {
	mock.Mock
}

func (m *mockJWTManager) GenerateTokenPair(userID, role string) (*jwt.TokenPair, error) {
	args := m.Called(userID, role)
	tokenPair, _ := args.Get(0).(*jwt.TokenPair)
	return tokenPair, args.Error(1)
}

func (m *mockJWTManager) ValidateRefreshToken(token string) (*jwt.Claims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(*jwt.Claims)
	return claims, args.Error(1)
}

func (m *mockJWTManager) RefreshTokenPair(refreshToken, userRole string) (*jwt.TokenPair, string, error) {
	args := m.Called(refreshToken, userRole)
	tokenPair, _ := args.Get(0).(*jwt.TokenPair)
	oldID, _ := args.Get(1).(string)
	return tokenPair, oldID, args.Error(2)
}

func (m *mockJWTManager) GetTokenID(token string) (string, error) {
	args := m.Called(token)
	id, _ := args.Get(0).(string)
	return id, args.Error(1)
}

type inMemorySessionStore struct {
	t    *testing.T
	mu   sync.Mutex
	data map[string][]byte
}

func newInMemorySessionStore(t *testing.T) *inMemorySessionStore {
	return &inMemorySessionStore{t: t, data: make(map[string][]byte)}
}

func (s *inMemorySessionStore) SetSession(_ context.Context, sessionID string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload, err := json.Marshal(data)
	require.NoError(s.t, err)
	s.data[sessionID] = payload
	return nil
}

func (s *inMemorySessionStore) GetSession(_ context.Context, sessionID string, dest interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.data[sessionID]
	if !ok {
		return errors.New("session not found")
	}

	return json.Unmarshal(value, dest)
}

func (s *inMemorySessionStore) DeleteSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[sessionID]; !ok {
		return errors.New("session not found")
	}
	delete(s.data, sessionID)
	return nil
}

func setupService(t *testing.T) (*authService, *mockAuthRepository, *mockJWTManager, *inMemorySessionStore, *inMemorySessionStore, *inMemorySessionStore) {
	t.Helper()

	repo := &mockAuthRepository{}
	jwtMgr := &mockJWTManager{}
	refreshStore := newInMemorySessionStore(t)
	resetStore := newInMemorySessionStore(t)
	verifyStore := newInMemorySessionStore(t)

	svcIface, err := NewAuthServiceWithStores(repo, jwtMgr, refreshStore, resetStore, verifyStore, Config{}, nil)
	require.NoError(t, err)

	svc := svcIface.(*authService)
	fixedTime := time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC)
	svc.now = func() time.Time { return fixedTime }

	return svc, repo, jwtMgr, refreshStore, resetStore, verifyStore
}

func TestRegisterSuccess(t *testing.T) {
	svc, repo, jwtMgr, refreshStore, _, _ := setupService(t)

	input := RegisterInput{
		Email:     "user@example.com",
		Password:  "Password123",
		FirstName: "Jane",
		LastName:  "Doe",
		UserType:  "volunteer",
		SecurityQuestions: []SecurityQuestionInput{
			{Question: "Q1", Answer: "A1"},
			{Question: "Q2", Answer: "A2"},
			{Question: "Q3", Answer: "A3"},
		},
	}

	repo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "user@example.com" && u.FirstName == "Jane"
	})).Return(nil).Once()

	tokenPair := &jwt.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	jwtMgr.On("GenerateTokenPair", mock.Anything, "volunteer").Return(tokenPair, nil).Once()
	claims := &jwt.Claims{RegisteredClaims: jwtlib.RegisteredClaims{ID: "refresh-id", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}}
	jwtMgr.On("ValidateRefreshToken", "refresh").Return(claims, nil).Once()

	resp, err := svc.Register(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access", resp.TokenPair.AccessToken)
	assert.Equal(t, "refresh", resp.TokenPair.RefreshToken)

	var stored refreshSessionData
	err = refreshStore.GetSession(context.Background(), "refresh-id", &stored)
	require.NoError(t, err)
	assert.Equal(t, "volunteer", stored.UserRole)

	repo.AssertExpectations(t)
	jwtMgr.AssertExpectations(t)
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc, repo, _, _, _, _ := setupService(t)

	repo.On("CreateUser", mock.Anything, mock.Anything).Return(repositories.ErrUserAlreadyExists).Once()

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:     "taken@example.com",
		Password:  "Password123",
		FirstName: "Jane",
		LastName:  "Doe",
		UserType:  "volunteer",
		SecurityQuestions: []SecurityQuestionInput{
			{Question: "Q1", Answer: "A1"},
			{Question: "Q2", Answer: "A2"},
			{Question: "Q3", Answer: "A3"},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
}

func TestLoginInvalidPassword(t *testing.T) {
	svc, repo, _, _, _, _ := setupService(t)

	user := &models.User{ID: uuid.New(), Email: "user@example.com", AccountStatus: models.AccountStatusActive}
	_ = user.SetPassword("Password123")

	repo.On("FindUserByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()

	_, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "wrong"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestLoginReactivatesInactiveAccount(t *testing.T) {
	svc, repo, jwtMgr, refreshStore, _, _ := setupService(t)

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "user@example.com", AccountStatus: models.AccountStatusInactive}
	require.NoError(t, user.SetPassword("Password123"))

	repo.On("FindUserByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()
	repo.On("UpdateAccountStatus", mock.Anything, userID, models.AccountStatusActive).Return(nil).Once()
	repo.On("UpdateLastLogin", mock.Anything, userID).Return(nil).Once()

	tokenPair := &jwt.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	jwtMgr.On("GenerateTokenPair", userID.String(), "volunteer").Return(tokenPair, nil).Once()
	claims := &jwt.Claims{RegisteredClaims: jwtlib.RegisteredClaims{ID: "refresh-id-login", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}}
	jwtMgr.On("ValidateRefreshToken", "refresh").Return(claims, nil).Once()

	resp, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "Password123"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, models.AccountStatusActive, resp.User.AccountStatus)

	var stored refreshSessionData
	err = refreshStore.GetSession(context.Background(), "refresh-id-login", &stored)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), stored.UserID)

	repo.AssertExpectations(t)
	jwtMgr.AssertExpectations(t)
}

func TestRefreshTokenSuccess(t *testing.T) {
	svc, _, jwtMgr, refreshStore, _, _ := setupService(t)

	session := refreshSessionData{UserID: uuid.New().String(), UserRole: "volunteer", TokenID: "old-id"}
	require.NoError(t, refreshStore.SetSession(context.Background(), "old-id", &session))

	jwtMgr.On("GetTokenID", "old-refresh").Return("old-id", nil).Once()
	jwtMgr.On("RefreshTokenPair", "old-refresh", "volunteer").Return(&jwt.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}, "old-id", nil).Once()
	claims := &jwt.Claims{RegisteredClaims: jwtlib.RegisteredClaims{ID: "new-id", ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}}
	jwtMgr.On("ValidateRefreshToken", "new-refresh").Return(claims, nil).Once()

	tokenPair, err := svc.RefreshToken(context.Background(), "old-refresh")
	require.NoError(t, err)
	assert.Equal(t, "new-access", tokenPair.AccessToken)

	var stored refreshSessionData
	require.NoError(t, refreshStore.GetSession(context.Background(), "new-id", &stored))
	assert.Equal(t, "volunteer", stored.UserRole)

	jwtMgr.AssertExpectations(t)
}

func TestLogoutInvalidToken(t *testing.T) {
	svc, _, jwtMgr, _, _, _ := setupService(t)

	jwtMgr.On("GetTokenID", "bogus").Return("", errors.New("invalid"))

	err := svc.Logout(context.Background(), "bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPasswordResetFlow(t *testing.T) {
	svc, repo, jwtMgr, _, _, verifyStore := setupService(t)

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "user@example.com", AccountStatus: models.AccountStatusActive}
	require.NoError(t, user.SetPassword("Password123"))
	require.NoError(t, user.SetSecurityAnswers("Q1", "A1", "Q2", "A2", "Q3", "A3"))

	repo.On("FindUserByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()

	challenge, err := svc.RequestPasswordReset(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, challenge.ResetToken)

	repo.On("FindUserByID", mock.Anything, userID).Return(user, nil).Once()

	answers := []string{"A1", "A2", "wrong"}
	verification, err := svc.VerifySecurityAnswers(context.Background(), challenge.ResetToken, answers)
	require.NoError(t, err)
	require.NotEmpty(t, verification.VerifiedToken)

	repo.On("UpdatePassword", mock.Anything, userID, mock.Anything).Return(nil).Once()

	err = svc.ConfirmPasswordReset(context.Background(), verification.VerifiedToken, "NewPassword123")
	require.NoError(t, err)

	err = verifyStore.GetSession(context.Background(), verification.VerifiedToken, &passwordResetVerificationData{})
	require.Error(t, err) // verification session deleted

	jwtMgr.AssertExpectations(t)
}
