package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	service "github.com/suhrobdomoiZ/Eda-1/services/courier/internal/services"
)

type CourierHandler struct {
	pb.UnimplementedCourierAPIServer
	svc *service.CourierService
}

func NewCourierHandler(svc *service.CourierService) *CourierHandler {
	return &CourierHandler{svc: svc}
}

// Посмотреть доступные заказы
func (h *CourierHandler) GetAvailableOrders(ctx context.Context, req *pb.GetAvailableOrdersRequest) (*pb.GetAvailableOrdersResponse, error) {
	courierID, ok := ctx.Value("user_id").(string)
	if !ok || courierID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	result, err := h.svc.GetAvailableOrders(ctx, courierID, limit)
	if err != nil {
		return nil, err
	}

	return &pb.GetAvailableOrdersResponse{
		Orders: result.Orders,
	}, nil
}

// Принять заказ
func (h *CourierHandler) AcceptOrder(ctx context.Context, req *pb.AcceptOrderRequest) (*pb.AcceptOrderResponse, error) {
	courierID, ok := ctx.Value("user_id").(string)
	if !ok || courierID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.AcceptOrder(ctx, courierID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.AcceptOrderResponse{
		Success: result.Success,
		Order:   result.Order,
	}, nil
}

// Все заказы курьера
func (h *CourierHandler) GetMyOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	// courier_id берём из запроса, если пусто — из контекста
	courierID := req.CourierId
	if courierID == "" {
		var ok bool
		courierID, ok = ctx.Value("user_id").(string)
		if !ok || courierID == "" {
			return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
		}
	}

	result, err := h.svc.GetMyOrders(ctx, courierID)
	if err != nil {
		return nil, err
	}

	return &pb.GetMyOrdersResponse{
		Orders: result.Orders,
	}, nil
}

// Забрать заказ
func (h *CourierHandler) PickUpOrder(ctx context.Context, req *pb.PickUpOrderRequest) (*pb.PickUpOrderResponse, error) {
	courierID, ok := ctx.Value("user_id").(string)
	if !ok || courierID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.PickUpOrder(ctx, courierID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.PickUpOrderResponse{
		Success: result.Success,
		Status:  result.Status,
	}, nil
}

// Сдать заказ
func (h *CourierHandler) DeliverOrder(ctx context.Context, req *pb.DeliverOrderRequest) (*pb.DeliverOrderResponse, error) {
	courierID, ok := ctx.Value("user_id").(string)
	if !ok || courierID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	result, err := h.svc.DeliverOrder(ctx, courierID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.DeliverOrderResponse{
		Success:  result.Success,
		Status:   result.Status,
		Earnings: result.Earnings,
	}, nil
}
