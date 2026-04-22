package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/auth"
	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/handlers"
)

type mockAuthClient struct {
	mock.Mock
}

func (m *mockAuthClient) Register(ctx context.Context, in *authpb.RegisterRequest, opts ...grpc.CallOption) (*authpb.RegisterResponse, error) {
	args := m.Called(in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.RegisterResponse), args.Error(1)
}

func (m *mockAuthClient) Login(ctx context.Context, in *authpb.LoginRequest, opts ...grpc.CallOption) (*authpb.LoginResponse, error) {
	args := m.Called(in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.LoginResponse), args.Error(1)
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, in *authpb.ValidateTokenRequest, opts ...grpc.CallOption) (*authpb.ValidateTokenResponse, error) {
	args := m.Called(in.AccessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.ValidateTokenResponse), args.Error(1)
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, in *authpb.RefreshTokenRequest, opts ...grpc.CallOption) (*authpb.RefreshTokenResponse, error) {
	args := m.Called(in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.RefreshTokenResponse), args.Error(1)
}

func (m *mockAuthClient) Logout(ctx context.Context, in *authpb.LogoutRequest, opts ...grpc.CallOption) (*authpb.LogoutResponse, error) {
	args := m.Called(in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.LogoutResponse), args.Error(1)
}

func (m *mockAuthClient) GetProfile(ctx context.Context, in *authpb.GetProfileRequest, opts ...grpc.CallOption) (*authpb.GetProfileResponse, error) {
	args := m.Called(in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.GetProfileResponse), args.Error(1)
}

// helpers

func newTestApp(client authpb.AuthServiceClient) *fiber.App {
	app := fiber.New()
	h := handlers.NewAuthHandler(client)
	app.Post("/register", h.Register)
	app.Post("/login", h.Login)
	app.Post("/refresh", h.Refresh)
	app.Post("/logout", h.Logout)
	app.Get("/profile", h.Profile)
	return app
}

func doRequest(app *fiber.App, method, path string, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	return resp
}

// Register

func TestRegisterHandler_Success(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("Register", mock.MatchedBy(func(r *authpb.RegisterRequest) bool {
		return r.Username == "alice" && r.Password == "pass123"
	})).Return(&authpb.RegisterResponse{
		UserId: "uuid-1",
		Tokens: &authpb.TokenPair{
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
	}, nil)

	resp := doRequest(app, "POST", "/register", map[string]string{
		"username": "alice",
		"password": "pass123",
		"role":     "customer",
	})

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "uuid-1", body["user_id"])
	assert.Equal(t, "access", body["access_token"])
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	resp := doRequest(app, "POST", "/register", map[string]string{
		"username": "",
		"password": "",
	})

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	client.AssertNotCalled(t, "Register")
}

func TestRegisterHandler_AlreadyExists(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("Register", mock.Anything).Return(nil,
		status.Error(codes.AlreadyExists, "username already taken"))

	resp := doRequest(app, "POST", "/register", map[string]string{
		"username": "alice",
		"password": "pass123",
		"role":     "customer",
	})

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
}

func TestLoginHandler_Success(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("Login", mock.MatchedBy(func(r *authpb.LoginRequest) bool {
		return r.Username == "bob" && r.Password == "secret"
	})).Return(&authpb.LoginResponse{
		UserId: "uuid-2",
		Role:   commonpb.UserRole_USER_ROLE_CUSTOMER,
		Tokens: &authpb.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
	}, nil)

	resp := doRequest(app, "POST", "/login", map[string]string{
		"username": "bob",
		"password": "secret",
	})

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "uuid-2", body["user_id"])
}

func TestLoginHandler_WrongCredentials(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("Login", mock.Anything).Return(nil,
		status.Error(codes.Unauthenticated, "invalid credentials"))

	resp := doRequest(app, "POST", "/login", map[string]string{
		"username": "bob",
		"password": "wrong",
	})

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRefreshHandler_Success(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("RefreshToken", mock.Anything).Return(&authpb.RefreshTokenResponse{
		Tokens: &authpb.TokenPair{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
		},
	}, nil)

	resp := doRequest(app, "POST", "/refresh", map[string]string{
		"refresh_token": "old-refresh",
	})

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "new-access", body["access_token"])
}

func TestRefreshHandler_InvalidToken(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("RefreshToken", mock.Anything).Return(nil,
		status.Error(codes.Unauthenticated, "refresh token invalid"))

	resp := doRequest(app, "POST", "/refresh", map[string]string{
		"refresh_token": "bad-token",
	})

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestLogoutHandler_Success(t *testing.T) {
	client := &mockAuthClient{}
	app := newTestApp(client)

	client.On("Logout", mock.Anything).Return(&authpb.LogoutResponse{Success: true}, nil)

	resp := doRequest(app, "POST", "/logout", map[string]string{
		"refresh_token": "some-token",
	})

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
