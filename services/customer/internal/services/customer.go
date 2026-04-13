package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	common "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/api"
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

	pgRepo *repository.PostgresRepo
}

func NewCustomerService(pgRepo *repository.PostgresRepo) *CustomerService {
	return &CustomerService{
		pgRepo: pgRepo,
	}
}

// Создание заказа

type CreateOrderInput struct {
	UserID       string
	RestaurantID string
	Items        []CreateOrderItemInput
	Address      string
}

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int32
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

	// TODO: Получить цены и названия товаров из сервиса ресторанов

	orderID := uuid.New().String()

	// Считаем общую сумму и формируем позиции
	var totalPrice int64
	var orderItems []repository.OrderItem

	for _, item := range input.Items {
		// TODO: Получить реальную цену из меню ресторана
		price := int64(50000) // В копейках

		totalPrice += price * int64(item.Quantity)

		orderItems = append(orderItems, repository.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Name:      "Блюдо", // TODO: Взять из меню
			Quantity:  item.Quantity,
			Price:     price,
		})
	}

	// Создаём заказ в БД
	order := &repository.Order{
		ID:           orderID,
		UserID:       input.UserID,
		RestaurantID: input.RestaurantID,
		Address:      input.Address,
		TotalPrice:   totalPrice,
		Status:       "created",
	}

	if err := s.pgRepo.CreateOrder(ctx, order, orderItems); err != nil {
		return nil, status.Errorf(codes.Internal, "create order: %v", err)
	}

	// TODO: Отправить событие в Kafka / уведомить ресторан

	return &CreateOrderResult{
		OrderID: orderID,
		Status:  "created",
	}, nil
}

// Получение заказа

type GetOrderResult struct {
	Order *common.Order
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
	pbOrder := &common.Order{
		Id:           orderWithItems.Order.ID,
		RestaurantId: orderWithItems.Order.RestaurantID,
		CourierId:    orderWithItems.Order.CourierID.String,
		ClientId:     orderWithItems.Order.UserID,
		Address:      orderWithItems.Order.Address,
		TotalPrice:   orderWithItems.Order.TotalPrice,
		Status:       s.MapStatus(orderWithItems.Order.Status),
	}

	for _, item := range orderWithItems.Items {
		pbOrder.Items = append(pbOrder.Items, &common.OrderItem{
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

	refundAmount := order.TotalPrice

	// TODO: Уведомить ресторан об отмене

	return &CancelOrderResult{
		Success:      true,
		RefundAmount: refundAmount,
	}, nil
}

func (s *CustomerService) canCancel(status string) bool {
	// Можно отменить только если заказ ещё не начали готовить
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
			Status:     s.MapStatus(o.Status),
			TotalPrice: o.TotalPrice,
		})
	}

	return &ListMyOrdersResult{
		Orders: pbOrders,
		Total:  total,
	}, nil
}

// Список ресторанов

type ListRestaurantsResult struct {
	Restaurants []*pb.RestaurantInfo
}

func (s *CustomerService) ListRestaurants(ctx context.Context, limit, offset int32) (*ListRestaurantsResult, error) {
	// TODO: Вызвать gRPC метод сервиса ресторанов

	return &ListRestaurantsResult{
		Restaurants: []*pb.RestaurantInfo{
			{Id: "rest_1", Name: "Пицца Миа"},
			{Id: "rest_2", Name: "Суши Мастер"},
			{Id: "rest_3", Name: "Бургер Кинг"},
		},
	}, nil
}

// Получение меню

type GetRestaurantMenuResult struct {
	RestaurantID   string
	RestaurantName string
	Items          []*pb.MenuItem
}

func (s *CustomerService) GetRestaurantMenu(ctx context.Context, restaurantID string) (*GetRestaurantMenuResult, error) {
	// TODO: Вызвать gRPC метод сервиса ресторанов

	return &GetRestaurantMenuResult{
		RestaurantID:   restaurantID,
		RestaurantName: s.getRestaurantName(ctx, restaurantID),
		Items: []*pb.MenuItem{
			{Id: "prod_1", Name: "Маргарита", Description: "Томаты, моцарелла, базилик", Price: 50000},
			{Id: "prod_2", Name: "Пепперони", Description: "Пепперони, моцарелла, томатный соус", Price: 60000},
			{Id: "prod_3", Name: "Кола", Description: "0.5л", Price: 10000},
		},
	}, nil
}

// Преобразование строкового статуса из БД в enum
func (s *CustomerService) MapStatus(status string) common.OrderStatus {
	switch status {
	case "created":
		return common.OrderStatus_ORDER_STATUS_CREATED
	case "cooking":
		return common.OrderStatus_ORDER_STATUS_COOKING
	case "ready":
		return common.OrderStatus_ORDER_STATUS_READY
	case "delivering":
		return common.OrderStatus_ORDER_STATUS_DELIVERING
	case "delivered":
		return common.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return common.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return common.OrderStatus_ORDER_STATUS_CREATED
	}
}

// Вспомогательные методы

// Получение названия ресторана
func (s *CustomerService) getRestaurantName(ctx context.Context, restaurantID string) string {
	// TODO: Вызвать gRPC метод сервиса ресторанов или взять из кэша
	mockNames := map[string]string{
		"rest_1": "Пицца Миа",
		"rest_2": "Суши Мастер",
		"rest_3": "Бургер Кинг",
	}

	if name, ok := mockNames[restaurantID]; ok {
		return name
	}
	return "Ресторан"
}
