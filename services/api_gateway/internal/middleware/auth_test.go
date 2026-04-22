package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/auth"
	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

type mockAuthClient struct {
	mock.Mock
}

func (m *mockAuthClient) ValidateToken(ctx interface{}, req *authpb.ValidateTokenRequest, opts ...interface{}) (*authpb.ValidateTokenResponse, error) {
	args := m.Called(req.AccessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.ValidateTokenResponse), args.Error(1)
}

func newMiddlewareApp(client authpb.AuthServiceClient) *fiber.App {
	app := fiber.New()
	app.Get("/protected", middleware.Auth(client), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": middleware.GetUserID(c),
			"role":    middleware.GetRole(c).String(),
		})
	})
	return app
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	client := &mockAuthClient{}
	app := newMiddlewareApp(client)

	client.On("ValidateToken", "valid-token").Return(&authpb.ValidateTokenResponse{
		Valid: true,
		Claims: &authpb.TokenClaims{
			UserId: "user-123",
			Role:   commonpb.UserRole_USER_ROLE_CUSTOMER,
		},
	}, nil)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	client := &mockAuthClient{}
	app := newMiddlewareApp(client)

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	client.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	client := &mockAuthClient{}
	app := newMiddlewareApp(client)

	client.On("ValidateToken", "bad-token").Return(&authpb.ValidateTokenResponse{
		Valid: false,
	}, nil)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_WrongScheme(t *testing.T) {
	client := &mockAuthClient{}
	app := newMiddlewareApp(client)

	// Basic вместо Bearer
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	client.AssertNotCalled(t, "ValidateToken")
}

func TestRequireRole_Correct(t *testing.T) {
	app := fiber.New()
	app.Get("/restaurant-only",
		func(c *fiber.Ctx) error {
			// симулируем Auth middleware — кладём роль вручную
			c.Locals("role", commonpb.UserRole_USER_ROLE_RESTAURANT)
			return c.Next()
		},
		middleware.RequireRole(commonpb.UserRole_USER_ROLE_RESTAURANT),
		func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) },
	)

	req := httptest.NewRequest("GET", "/restaurant-only", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireRole_WrongRole(t *testing.T) {
	app := fiber.New()
	app.Get("/restaurant-only",
		func(c *fiber.Ctx) error {
			c.Locals("role", commonpb.UserRole_USER_ROLE_CUSTOMER) // клиент
			return c.Next()
		},
		middleware.RequireRole(commonpb.UserRole_USER_ROLE_RESTAURANT),
		func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) },
	)

	req := httptest.NewRequest("GET", "/restaurant-only", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
