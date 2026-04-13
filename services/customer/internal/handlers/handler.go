package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/api"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

type CustomerHandler struct {
	pb.UnimplementedCustomerAPIServer
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	items := make([]service.CreateOrderItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.CreateOrderItemInput{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	result, err := h.svc.CreateOrder(ctx, &service.CreateOrderInput{
		UserID:       userID,
		RestaurantID: req.RestaurantId,
		Items:        items,
		Address:      req.Address,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderId: result.OrderID,
		Status:  h.svc.MapStatus(result.Status),
	}, nil
}

func (h *CustomerHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.GetOrder(ctx, userID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderResponse{Order: result.Order}, nil
}

func (h *CustomerHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.CancelOrder(ctx, userID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.CancelOrderResponse{
		Success:      result.Success,
		RefundAmount: result.RefundAmount,
	}, nil
}

func (h *CustomerHandler) ListMyOrders(ctx context.Context, req *pb.ListMyOrdersRequest) (*pb.ListMyOrdersResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.ListMyOrders(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &pb.ListMyOrdersResponse{Orders: result.Orders}, nil
}

func (h *CustomerHandler) GetRestaurantMenu(ctx context.Context, req *pb.GetRestaurantMenuRequest) (*pb.GetRestaurantMenuResponse, error) {
	result, err := h.svc.GetRestaurantMenu(ctx, req.RestaurantId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRestaurantMenuResponse{
		RestaurantId:   result.RestaurantID,
		RestaurantName: result.RestaurantName,
	}, nil
}

func (h *CustomerHandler) ListRestaurants(ctx context.Context, req *pb.ListRestaurantsRequest) (*pb.ListRestaurantsResponse, error) {
	result, err := h.svc.ListRestaurants(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &pb.ListRestaurantsResponse{Restaurants: result.Restaurants}, nil
}
