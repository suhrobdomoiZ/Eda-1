package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	common_api "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	"github.com/suhrobdomoiZ/Eda-1/pkg/grpcmeta"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	service "github.com/suhrobdomoiZ/Eda-1/services/courier/internal/services"
)

func NewOrderConsumerHandler(svc *service.CourierService) kafka.HandlerFunc {
	return func(ctx context.Context, key string, value []byte) error {
		var event kafka.ChangeOrderStatusEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf("unmarshal event: %w", err)
		}

		switch event.NewStatus {
		case common_api.OrderStatus_ORDER_STATUS_READY:
			// TODO: уведомить курьеров поблизости
		}

		return nil
	}
}

type CourierHandler struct {
	pb.UnimplementedCourierAPIServer
	svc *service.CourierService
}

func NewCourierHandler(svc *service.CourierService) *CourierHandler {
	return &CourierHandler{svc: svc}
}

func (h *CourierHandler) GetAvailableOrders(ctx context.Context, req *pb.GetAvailableOrdersRequest) (*pb.GetAvailableOrdersResponse, error) {
	courierID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
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

func (h *CourierHandler) AcceptOrder(ctx context.Context, req *pb.AcceptOrderRequest) (*pb.AcceptOrderResponse, error) {
	courierID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
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

func (h *CourierHandler) GetMyOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	// courier_id всегда из metadata — поле в req игнорируем
	courierID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := h.svc.GetMyOrders(ctx, courierID)
	if err != nil {
		return nil, err
	}

	return &pb.GetMyOrdersResponse{
		Orders: result.Orders,
	}, nil
}

func (h *CourierHandler) PickUpOrder(ctx context.Context, req *pb.PickUpOrderRequest) (*pb.PickUpOrderResponse, error) {
	courierID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
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

func (h *CourierHandler) DeliverOrder(ctx context.Context, req *pb.DeliverOrderRequest) (*pb.DeliverOrderResponse, error) {
	courierID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
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
