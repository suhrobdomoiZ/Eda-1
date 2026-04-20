package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	authpb "github.com/suhrobdomoiZ/Eda-1/services/api"
)

type contextKey string

const (
	KeyUserID contextKey = "user_id"
	KeyRole   contextKey = "role"
)

func Auth(authClient authpb.AuthServiceClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearer(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		resp, err := authClient.ValidateToken(
			context.Background(),
			&authpb.ValidateTokenRequest{AccessToken: token},
		)
		if err != nil || !resp.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals(string(KeyUserID), resp.Claims.UserId)
		c.Locals(string(KeyRole), resp.Claims.Role)

		return c.Next()
	}
}

// RequireRole проверяет роль - ставится после Auth
func RequireRole(role commonpb.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(string(KeyRole)).(commonpb.UserRole)
		if !ok || userRole != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
		return c.Next()
	}
}

func extractBearer(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

// GetUserID достаёт user_id из fiber.Ctx - хелпер для хендлеров
func GetUserID(c *fiber.Ctx) string {
	id, _ := c.Locals(string(KeyUserID)).(string)
	return id
}

// GetRole достаёт role из fiber.Ctx
func GetRole(c *fiber.Ctx) commonpb.UserRole {
	role, _ := c.Locals(string(KeyRole)).(commonpb.UserRole)
	return role
}
