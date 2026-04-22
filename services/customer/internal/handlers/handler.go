package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	common_api "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	common_methods "github.com/suhrobdomoiZ/Eda-1/pkg/common_methods"
	"github.com/suhrobdomoiZ/Eda-1/pkg/grpcmeta"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

func NewOrderConsumerHandler(svc *service.CustomerService) kafka.HandlerFunc {
	return func(ctx context.Context, key string, value []byte) error {
		var event kafka.ChangeOrderStatusEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf("unmarshal event: %w", err)
		}

		switch event.NewStatus {
		case common_api.OrderStatus_ORDER_STATUS_DELIVERED:
			// TODO: push-уведомление клиенту
		case common_api.OrderStatus_ORDER_STATUS_CANCELLED:
			// TODO: уведомить клиента об отмене рестораном
		}

		return nil
	}
}

type CustomerHandler struct {
	pb.UnimplementedCustomerAPIServer
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	userID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := h.svc.CreateOrder(ctx, &service.CreateOrderInput{
		UserID:       userID,
		RestaurantID: req.RestaurantId,
		Items:        req.Items,
		Address:      req.Address,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderId: result.OrderID,
		Status:  common_methods.MapOrderStatus(result.Status),
	}, nil
}

func (h *CustomerHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	userID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := h.svc.GetOrder(ctx, userID, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderResponse{Order: result.Order}, nil
}

func (h *CustomerHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	userID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
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
	userID, err := grpcmeta.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := h.svc.ListMyOrders(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &pb.ListMyOrdersResponse{Orders: result.Orders}, nil
}
