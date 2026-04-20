package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	authpb "github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

type AuthHandler struct {
	client authpb.AuthServiceClient
}

func NewAuthHandler(client authpb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: client}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"` // customer, restaurant, courier

		RestaurantName    string `json:"restaurant_name"`
		RestaurantAddress string `json:"restaurant_address"`
		RestaurantPhone   string `json:"restaurant_phone"`

		CourierName  string `json:"courier_name"`
		CourierPhone string `json:"courier_phone"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password required"})
	}

	protoReq := &authpb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Role:     stringToRole(req.Role),
	}

	switch req.Role {
	case "restaurant":
		protoReq.Profile = &authpb.RegisterRequest_Restaurant{
			Restaurant: &authpb.RestaurantRegisterInfo{
				Name:    req.RestaurantName,
				Address: req.RestaurantAddress,
				Phone:   req.RestaurantPhone,
			},
		}
	case "courier":
		protoReq.Profile = &authpb.RegisterRequest_Courier{
			Courier: &authpb.CourierRegisterInfo{
				Name:  req.CourierName,
				Phone: req.CourierPhone,
			},
		}
	}

	resp, err := h.client.Register(context.Background(), protoReq)
	if err != nil {
		return grpcError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user_id":       resp.UserId,
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
	})
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.client.Login(context.Background(), &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return c.JSON(fiber.Map{
		"user_id":       resp.UserId,
		"role":          resp.Role.String(),
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
	})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.client.RefreshToken(context.Background(), &authpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return c.JSON(fiber.Map{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
	})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	_, err := h.client.Logout(context.Background(), &authpb.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return c.JSON(fiber.Map{"success": true})
}

// GET /api/v1/auth/profile  [JWT]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	resp, err := h.client.GetProfile(context.Background(), &authpb.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return grpcError(c, err)
	}

	result := fiber.Map{
		"id":       resp.User.Id,
		"username": resp.User.Username,
		"role":     resp.User.Role.String(),
	}

	switch ext := resp.Extended.(type) {
	case *authpb.GetProfileResponse_Restaurant:
		result["restaurant"] = fiber.Map{
			"name":    ext.Restaurant.Name,
			"address": ext.Restaurant.Address,
			"phone":   ext.Restaurant.Phone,
		}
	case *authpb.GetProfileResponse_Courier:
		result["courier"] = fiber.Map{
			"name":  ext.Courier.Name,
			"phone": ext.Courier.Phone,
		}
	}

	return c.JSON(result)
}

func stringToRole(r string) commonpb.UserRole {
	switch r {
	case "restaurant":
		return commonpb.UserRole_USER_ROLE_RESTAURANT
	case "courier":
		return commonpb.UserRole_USER_ROLE_COURIER
	case "admin":
		return commonpb.UserRole_USER_ROLE_ADMIN
	default:
		return commonpb.UserRole_USER_ROLE_CUSTOMER
	}
}
