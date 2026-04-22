package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	common_api "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	common_methods "github.com/suhrobdomoiZ/Eda-1/pkg/common_methods"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/repository"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderNotBelongsTo = errors.New("order does not belong to user")
	ErrOrderCannotCancel = errors.New("order cannot be cancelled")
	ErrInvalidInput      = errors.New("invalid input")
)

type CustomerService struct {
	pb.UnimplementedCustomerAPIServer
	pgRepo   *repository.PostgresRepo
	producer *kafka.Producer
	metrics  *metrics.Metrics
}

func NewCustomerService(
	pgRepo *repository.PostgresRepo,
	p *kafka.Producer,
	m *metrics.Metrics,
) *CustomerService {
	return &CustomerService{
		pgRepo:   pgRepo,
		producer: p,
		metrics:  m,
	}
}

// Создание заказа

type CreateOrderInput struct {
	UserID       string
	RestaurantID string
	Items        []*pb.CreateOrderItem
	Address      string
}

type CreateOrderResult struct {
	OrderID string
	Status  string
}

func (s *CustomerService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderResult, error) {
	if input.RestaurantID == "" {
		return nil, status.Error(codes.InvalidArgument, "restaurant_id is required")
	}
	if len(input.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items cannot be empty")
	}
	if input.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	orderID := uuid.New().String()
	var totalPrice int64
	var orderItems []repository.OrderItem

	for _, item := range input.Items {
		totalPrice += item.Price * int64(item.Quantity)

		orderItems = append(orderItems, repository.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductId,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	order := &repository.Order{
		ID:           orderID,
		UserID:       input.UserID,
		RestaurantID: input.RestaurantID,
		Address:      input.Address,
		TotalPrice:   totalPrice,
		Status:       "created",
	}

	if err := s.pgRepo.CreateOrder(ctx, order, orderItems); err != nil {
		// TODO: error metric
		return nil, status.Errorf(codes.Internal, "create order: %v", err)
	}

	// Kafka event
	event := kafka.ChangeOrderStatusEvent{
		OrderId:   uuid.MustParse(orderID),
		NewStatus: common_api.OrderStatus_ORDER_STATUS_CREATED,
	}
	_ = s.producer.Send(ctx, orderID, event)

	s.metrics.OnOrderCreated(float64(totalPrice / 100))

	return &CreateOrderResult{
		OrderID: orderID,
		Status:  "created",
	}, nil
}

// Получение заказа

type GetOrderResult struct {
	Order *common_api.Order
}

func (s *CustomerService) GetOrder(ctx context.Context, userID, orderID string) (*GetOrderResult, error) {
	belongs, err := s.pgRepo.CheckOrderBelongsToUser(ctx, orderID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check order belongs: %v", err)
	}
	if !belongs {
		return nil, status.Error(codes.PermissionDenied, "order does not belong to user")
	}

	// Получаем заказ с позициями
	orderWithItems, err := s.pgRepo.GetOrderWithItems(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}

	// Преобразуем в protobuf
	pbOrder := &common_api.Order{
		Id:           orderWithItems.Order.ID,
		RestaurantId: orderWithItems.Order.RestaurantID,
		CourierId:    orderWithItems.Order.CourierID.String,
		ClientId:     orderWithItems.Order.UserID,
		Address:      orderWithItems.Order.Address,
		TotalPrice:   orderWithItems.Order.TotalPrice,
		Status:       common_methods.MapOrderStatus(orderWithItems.Order.Status),
	}

	for _, item := range orderWithItems.Items {
		pbOrder.Items = append(pbOrder.Items, &common_api.OrderItem{
			ProductId: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return &GetOrderResult{Order: pbOrder}, nil
}

// Отмена заказа

type CancelOrderResult struct {
	Success      bool
	RefundAmount int64
}

func (s *CustomerService) CancelOrder(ctx context.Context, userID, orderID string) (*CancelOrderResult, error) {
	belongs, err := s.pgRepo.CheckOrderBelongsToUser(ctx, orderID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check order belongs: %v", err)
	}
	if !belongs {
		return nil, status.Error(codes.PermissionDenied, "order does not belong to user")
	}

	// Получаем заказ
	order, err := s.pgRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}

	// Проверяем, можно ли отменить
	if !s.canCancel(order.Status) {
		return nil, status.Error(codes.FailedPrecondition, "order cannot be cancelled at this stage")
	}

	// Отменяем заказ
	if err := s.pgRepo.CancelOrder(ctx, orderID); err != nil {
		return nil, status.Errorf(codes.Internal, "cancel order: %v", err)
	}

	// kafka event
	event := kafka.ChangeOrderStatusEvent{
		OrderId:   uuid.MustParse(orderID),
		NewStatus: common_api.OrderStatus_ORDER_STATUS_CANCELLED,
	}
	if err := s.producer.Send(ctx, orderID, event); err != nil {
		// TODO: error metric
	}

	s.metrics.OnOrderCancelled()

	refundAmount := order.TotalPrice

	return &CancelOrderResult{
		Success:      true,
		RefundAmount: refundAmount,
	}, nil
}

func (s *CustomerService) canCancel(status string) bool {
	cancellableStatuses := map[string]bool{
		"created": true,
		"cooking": true,
	}
	return cancellableStatuses[status]
}

// Список заказов пользователя

type ListMyOrdersResult struct {
	Orders []*pb.OrderInfo
	Total  int32
}

func (s *CustomerService) ListMyOrders(ctx context.Context, userID string, limit, offset int32) (*ListMyOrdersResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	orders, err := s.pgRepo.ListOrdersByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list orders: %v", err)
	}

	total, err := s.pgRepo.CountOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count orders: %v", err)
	}

	var pbOrders []*pb.OrderInfo
	for _, o := range orders {
		pbOrders = append(pbOrders, &pb.OrderInfo{
			Id:         o.ID,
			Status:     common_methods.MapOrderStatus(o.Status),
			TotalPrice: o.TotalPrice,
		})
	}

	return &ListMyOrdersResult{
		Orders: pbOrders,
		Total:  total,
	}, nil
}
