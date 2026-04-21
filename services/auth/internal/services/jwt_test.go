package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/services"
)

func newTestJWT() *services.JWTService {
	return services.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
}

func TestJWT_GenerateAndParse_AccessToken(t *testing.T) {
	jwt := newTestJWT()

	token, err := jwt.GenerateAccessToken("user-123", "user")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestJWT_GenerateAndParse_RefreshToken(t *testing.T) {
	jwt := newTestJWT()

	token, err := jwt.GenerateRefreshToken("user-456", "courier")
	require.NoError(t, err)

	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "courier", claims.Role)
}

func TestJWT_ParseToken_InvalidSignature(t *testing.T) {
	jwt1 := services.NewJWTService("secret-1", 15*time.Minute, 7*24*time.Hour)
	jwt2 := services.NewJWTService("secret-2", 15*time.Minute, 7*24*time.Hour)

	token, err := jwt1.GenerateAccessToken("user-1", "user")
	require.NoError(t, err)

	_, err = jwt2.ParseToken(token)
	assert.ErrorIs(t, err, services.ErrTokenInvalid)
}

func TestJWT_ParseToken_Expired(t *testing.T) {
	jwt := services.NewJWTService("test-secret", -time.Minute, -time.Minute)

	token, err := jwt.GenerateAccessToken("user-1", "user")
	require.NoError(t, err)

	_, err = jwt.ParseToken(token)
	assert.ErrorIs(t, err, services.ErrTokenExpired)
}

func TestJWT_ParseToken_Garbage(t *testing.T) {
	jwt := newTestJWT()

	_, err := jwt.ParseToken("not.a.token")
	assert.ErrorIs(t, err, services.ErrTokenInvalid)
}
