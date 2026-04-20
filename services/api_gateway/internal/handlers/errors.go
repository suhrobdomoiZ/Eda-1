package handlers

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcError(c *fiber.Ctx, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": st.Message()})
	case codes.NotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": st.Message()})
	case codes.AlreadyExists:
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": st.Message()})
	case codes.Unauthenticated:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": st.Message()})
	case codes.PermissionDenied:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": st.Message()})
	case codes.Unavailable:
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "service unavailable"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
}
