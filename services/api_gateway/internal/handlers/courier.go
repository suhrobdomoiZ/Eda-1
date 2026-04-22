package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/metadata"

	courierpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

type CourierHandler struct {
	client courierpb.CourierAPIClient
}

func NewCourierHandler(client courierpb.CourierAPIClient) *CourierHandler {
	return &CourierHandler{client: client}
}

func outgoingCtx(c *fiber.Ctx) context.Context {
	md := metadata.Pairs(
		"user_id", middleware.GetUserID(c),
		"role", middleware.GetRole(c).String(),
	)
	return metadata.NewOutgoingContext(context.Background(), md)
}

// GET /api/v1/courier/orders/available  [JWT, role=courier]
func (h *CourierHandler) GetAvailableOrders(c *fiber.Ctx) error {
	resp, err := h.client.GetAvailableOrders(outgoingCtx(c), &courierpb.GetAvailableOrdersRequest{
		Limit: int32(c.QueryInt("limit", 20)),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/accept  [JWT, role=courier]
func (h *CourierHandler) AcceptOrder(c *fiber.Ctx) error {
	resp, err := h.client.AcceptOrder(outgoingCtx(c), &courierpb.AcceptOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/courier/orders  [JWT, role=courier]
func (h *CourierHandler) GetMyOrders(c *fiber.Ctx) error {
	resp, err := h.client.GetMyOrders(outgoingCtx(c), &courierpb.GetMyOrdersRequest{})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/pickup  [JWT, role=courier]
func (h *CourierHandler) PickUpOrder(c *fiber.Ctx) error {
	resp, err := h.client.PickUpOrder(outgoingCtx(c), &courierpb.PickUpOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/deliver  [JWT, role=courier]
func (h *CourierHandler) DeliverOrder(c *fiber.Ctx) error {
	resp, err := h.client.DeliverOrder(outgoingCtx(c), &courierpb.DeliverOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}
