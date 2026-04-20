package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	rpb "github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

type RestaurantHandler struct {
	client rpb.RestaurantClient
}

func NewRestaurantHandler(client rpb.RestaurantClient) *RestaurantHandler {
	return &RestaurantHandler{client: client}
}

// GET /api/v1/restaurant/menu/:restaurant_id  [публичный]
func (h *RestaurantHandler) ListProducts(c *fiber.Ctx) error {
	resp, err := h.client.ListProducts(context.Background(), &rpb.ListProductsRequest{
		RestaurantId: c.Params("restaurant_id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/restaurant/menu/:restaurant_id/product/:id  [публичный]
func (h *RestaurantHandler) GetProduct(c *fiber.Ctx) error {
	resp, err := h.client.GetProduct(context.Background(), &rpb.GetProductRequest{
		Id: c.Params("id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// POST /api/v1/restaurant/menu  [JWT, role=restaurant]
func (h *RestaurantHandler) AddProduct(c *fiber.Ctx) error {
	var req rpb.ProductInfo
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	// restaurant_id берём из токена, не из тела запроса
	req.RestaurantId = middleware.GetUserID(c)

	resp, err := h.client.AddProduct(context.Background(), &rpb.AddProductRequest{
		ProductInfo: &req,
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// PUT /api/v1/restaurant/menu/:id  [JWT, role=restaurant]
func (h *RestaurantHandler) UpdateProduct(c *fiber.Ctx) error {
	var info rpb.ProductInfo
	if err := c.BodyParser(&info); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	info.RestaurantId = middleware.GetUserID(c)

	resp, err := h.client.UpdateProduct(context.Background(), &rpb.UpdateProductRequest{
		Id:          c.Params("id"),
		ProductInfo: &info,
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// DELETE /api/v1/restaurant/menu/:id  [JWT, role=restaurant]
func (h *RestaurantHandler) DeleteProduct(c *fiber.Ctx) error {
	resp, err := h.client.DeleteProduct(context.Background(), &rpb.DeleteProductRequest{
		Id: c.Params("id"),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// GET /api/v1/restaurant/orders  [JWT, role=restaurant]
func (h *RestaurantHandler) ListOrders(c *fiber.Ctx) error {
	resp, err := h.client.ListOrders(context.Background(), &rpb.ListOrdersRequest{
		Id: middleware.GetUserID(c),
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}

// PUT /api/v1/restaurant/orders/:id/status  [JWT, role=restaurant]
func (h *RestaurantHandler) ChangeOrderStatus(c *fiber.Ctx) error {
	var req struct {
		Status commonpb.OrderStatus `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.client.ChangeOrderStatus(context.Background(), &rpb.ChangeOrderStatusRequest{
		Id:     c.Params("id"),
		Status: req.Status,
	})
	if err != nil {
		return grpcError(c, err)
	}
	return c.JSON(resp)
}
