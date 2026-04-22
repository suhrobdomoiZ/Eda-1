package handlers

import (
	"github.com/gofiber/fiber/v2"

	cpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
)

type CustomerHandler struct {
	client cpb.CustomerAPIClient
}

func NewCustomerHandler(client cpb.CustomerAPIClient) *CustomerHandler {
	return &CustomerHandler{client: client}
}

// POST /api/v1/customer/orders  [JWT, role=customer]
func (h *CustomerHandler) CreateOrder(c *fiber.Ctx) error {
	var req cpb.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.client.CreateOrder(outgoingCtx(c), &req)
	if err != nil {
		return grpcError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GET /api/v1/customer/orders  [JWT, role=customer]
func (h *CustomerHandler) ListMyOrders(c *fiber.Ctx) error {
	resp, err := h.client.ListMyOrders(outgoingCtx(c), &cpb.ListMyOrdersRequest{
		Limit:  int32(c.QueryInt("limit", 20)),
		Offset: int32(c.QueryInt("offset", 0)),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/customer/orders/:order_id  [JWT, role=customer]
func (h *CustomerHandler) GetOrder(c *fiber.Ctx) error {
	resp, err := h.client.GetOrder(outgoingCtx(c), &cpb.GetOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// DELETE /api/v1/customer/orders/:order_id  [JWT, role=customer]
func (h *CustomerHandler) CancelOrder(c *fiber.Ctx) error {
	resp, err := h.client.CancelOrder(outgoingCtx(c), &cpb.CancelOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}
