package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/repository"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/services"
)

type mockMetrics struct {
	mock.Mock
}

func (m *mockMetrics) IncError(method, errorType string) {
	m.Called(method, errorType)
}

func (m *mockMetrics) IncRequest(method, status string) {
	m.Called(method, status)
}

func (m *mockMetrics) ObserveDuration(method string, duration float64) {
	m.Called(method, duration)
}

func newTestAuthService(pg *mockPgRepo, rdb *mockRedisRepo, m *mockMetrics) *services.AuthService {
	jwt := services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	return services.NewAuthService(pg, rdb, jwt, m)
}

// Register

func TestRegister_Customer_Success(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	pg.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *repository.User) bool {
		return u.Username == "alice" && u.Role == "user"
	})).Return(nil)

	rdb.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("string"),
		mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	result, err := svc.Register(context.Background(), services.RegisterInput{
		Username: "alice",
		Password: "password123",
		Role:     "user",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	pg.AssertExpectations(t)
	rdb.AssertExpectations(t)
}

func TestRegister_Restaurant_CreatesProfile(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	pg.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *repository.User) bool {
		return u.Role == "restaurant"
	})).Return(nil)

	pg.On("CreateRestaurantProfile", mock.Anything, mock.MatchedBy(func(p *repository.RestaurantProfile) bool {
		return p.Name == "Pizza Place" && p.Phone == "+7 777 000 00 00"
	})).Return(nil)

	rdb.On("SaveRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Register(context.Background(), services.RegisterInput{
		Username:          "pizza",
		Password:          "secret",
		Role:              "restaurant",
		RestaurantName:    "Pizza Place",
		RestaurantAddress: "ул. Ленина 1",
		RestaurantPhone:   "+7 777 000 00 00",
	})

	require.NoError(t, err)
	pg.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	pg.On("CreateUser", mock.Anything, mock.Anything).Return(repository.ErrAlreadyExists)

	_, err := svc.Register(context.Background(), services.RegisterInput{
		Username: "alice",
		Password: "password123",
		Role:     "user",
	})

	assert.ErrorIs(t, err, services.ErrUserAlreadyExists)
	rdb.AssertNotCalled(t, "SaveRefreshToken")
}

// Login

func TestLogin_Success(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	hash, _ := bcryptHash("secret")

	pg.On("GetUserByUsername", mock.Anything, "bob").Return(&repository.User{
		ID:           "user-bob",
		Username:     "bob",
		PasswordHash: hash,
		Role:         "user",
	}, nil)

	rdb.On("SaveRefreshToken", mock.Anything, "user-bob",
		mock.AnythingOfType("string"), mock.Anything).Return(nil)

	result, err := svc.Login(context.Background(), "bob", "secret")

	require.NoError(t, err)
	assert.Equal(t, "user-bob", result.UserID)
	assert.Equal(t, "user", result.Role)
	assert.NotEmpty(t, result.AccessToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	hash, _ := bcryptHash("correct-password")

	pg.On("GetUserByUsername", mock.Anything, "bob").Return(&repository.User{
		ID:           "user-bob",
		Username:     "bob",
		PasswordHash: hash,
		Role:         "user",
	}, nil)

	_, err := svc.Login(context.Background(), "bob", "wrong-password")
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	rdb.AssertNotCalled(t, "SaveRefreshToken")
}

func TestLogin_UserNotFound(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	pg.On("GetUserByUsername", mock.Anything, "ghost").Return(nil, repository.ErrNotFound)

	_, err := svc.Login(context.Background(), "ghost", "any")
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

// RefreshToken

func TestRefreshToken_Success(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	jwt := services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	metrics := &mockMetrics{}
	svc := services.NewAuthService(pg, rdb, jwt, metrics)

	refreshToken, _ := jwt.GenerateRefreshToken("user-1", "user")

	rdb.On("GetUserIDByRefreshToken", mock.Anything, refreshToken).Return("user-1", nil)
	rdb.On("DeleteRefreshToken", mock.Anything, "user-1", refreshToken).Return(nil)
	rdb.On("SaveRefreshToken", mock.Anything, "user-1",
		mock.AnythingOfType("string"), mock.Anything).Return(nil)

	result, err := svc.RefreshToken(context.Background(), refreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEqual(t, refreshToken, result.RefreshToken)
	rdb.AssertExpectations(t)
}

func TestRefreshToken_NotInRedis(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	jwt := services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	metrics := &mockMetrics{}
	svc := services.NewAuthService(pg, rdb, jwt, metrics)

	refreshToken, _ := jwt.GenerateRefreshToken("user-1", "user")

	rdb.On("GetUserIDByRefreshToken", mock.Anything, refreshToken).Return("", repository.ErrNotFound)

	_, err := svc.RefreshToken(context.Background(), refreshToken)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

// Logout

func TestLogout_Success(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	jwt := services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	metrics := &mockMetrics{}
	svc := services.NewAuthService(pg, rdb, jwt, metrics)

	refreshToken, _ := jwt.GenerateRefreshToken("user-1", "user")
	rdb.On("DeleteRefreshToken", mock.Anything, "user-1", refreshToken).Return(nil)

	err := svc.Logout(context.Background(), refreshToken)
	require.NoError(t, err)
	rdb.AssertExpectations(t)
}

func TestLogout_ExpiredToken_NoError(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}

	jwt := services.NewJWTService("test-secret", -time.Minute, -time.Minute)
	metrics := &mockMetrics{}
	svc := services.NewAuthService(pg, rdb, jwt, metrics)

	expiredToken, _ := jwt.GenerateRefreshToken("user-1", "user")

	err := svc.Logout(context.Background(), expiredToken)
	require.NoError(t, err)
	rdb.AssertNotCalled(t, "DeleteRefreshToken")
}

// ValidateToken

func TestValidateToken_Valid(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	jwtSvc := services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	token, _ := jwtSvc.GenerateAccessToken("user-1", "courier")

	claims, err := svc.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "courier", claims.Role)
}

func TestValidateToken_Invalid(t *testing.T) {
	pg := &mockPgRepo{}
	rdb := &mockRedisRepo{}
	metrics := &mockMetrics{}
	svc := newTestAuthService(pg, rdb, metrics)

	metrics.On("IncError", "validate_token", "validation").Return()

	_, err := svc.ValidateToken(context.Background(), "garbage-token")
	assert.Error(t, err)

	metrics.AssertCalled(t, "IncError", "validate_token", "validation")
}

// helpers

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
