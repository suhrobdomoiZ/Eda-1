package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/metadata"

	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

func outgoingCtx(c *fiber.Ctx) context.Context {
	md := metadata.Pairs(
		"user_id", middleware.GetUserID(c),
		"role", middleware.GetRole(c).String(),
	)
	return metadata.NewOutgoingContext(context.Background(), md)
}
