package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	cpb "github.com/suhrobdomoiZ/Eda-1/services/api"
)

type CustomerHandler struct {
	client cpb.CustomerAPIClient
}

func NewCustomerHandler(client cpb.CustomerAPIClient) *CustomerHandler {
	return &CustomerHandler{client: client}
}

// GET /api/v1/customer/restaurants  [публичный]
func (h *CustomerHandler) ListRestaurants(c *fiber.Ctx) error {
	resp, err := h.client.ListRestaurants(context.Background(), &cpb.ListRestaurantsRequest{
		Limit:  int32(c.QueryInt("limit", 20)),
		Offset: int32(c.QueryInt("offset", 0)),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/customer/restaurants/:restaurant_id/menu  [публичный]
func (h *CustomerHandler) GetRestaurantMenu(c *fiber.Ctx) error {
	resp, err := h.client.GetRestaurantMenu(context.Background(), &cpb.GetRestaurantMenuRequest{
		RestaurantId: c.Params("restaurant_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/customer/orders  [JWT, role=customer]
func (h *CustomerHandler) CreateOrder(c *fiber.Ctx) error {
	var req cpb.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.client.CreateOrder(context.Background(), &req)
	if err != nil {
		return grpcError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GET /api/v1/customer/orders  [JWT, role=customer]
func (h *CustomerHandler) ListMyOrders(c *fiber.Ctx) error {
	resp, err := h.client.ListMyOrders(context.Background(), &cpb.ListMyOrdersRequest{
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
	resp, err := h.client.GetOrder(context.Background(), &cpb.GetOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// DELETE /api/v1/customer/orders/:order_id  [JWT, role=customer]
func (h *CustomerHandler) CancelOrder(c *fiber.Ctx) error {
	resp, err := h.client.CancelOrder(context.Background(), &cpb.CancelOrderRequest{
		OrderId: c.Params("order_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}
