package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	courierpb "github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

type CourierHandler struct {
	client courierpb.ClientAPIClient
}

func NewCourierHandler(client courierpb.ClientAPIClient) *CourierHandler {
	return &CourierHandler{client: client}
}

// GET /api/v1/courier/orders/available  [JWT, role=courier]
func (h *CourierHandler) GetAvailableOrders(c *fiber.Ctx) error {
	resp, err := h.client.GetAvailableOrders(context.Background(), &courierpb.GetAvailableOrdersRequest{
		Limit: int32(c.QueryInt("limit", 20)),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/accept  [JWT, role=courier]
func (h *CourierHandler) AcceptOrder(c *fiber.Ctx) error {
	resp, err := h.client.AcceptOrder(context.Background(), &courierpb.AcceptOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/courier/orders  [JWT, role=courier]
func (h *CourierHandler) GetMyOrders(c *fiber.Ctx) error {
	// courier_id берём из токена — не из запроса
	resp, err := h.client.GetMyOrders(context.Background(), &courierpb.GetMyOrdersRequest{
		CourierId: middleware.GetUserID(c),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/pickup  [JWT, role=courier]
func (h *CourierHandler) PickUpOrder(c *fiber.Ctx) error {
	resp, err := h.client.PickUpOrder(context.Background(), &courierpb.PickUpOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/courier/orders/:order_id/deliver  [JWT, role=courier]
func (h *CourierHandler) DeliverOrder(c *fiber.Ctx) error {
	resp, err := h.client.DeliverOrder(context.Background(), &courierpb.DeliverOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}
