package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	common_methods "github.com/suhrobdomoiZ/Eda-1/pkg/common_methods"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/repository"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderNotAvailable = errors.New("order is not available")
	ErrOrderAlreadyTaken = errors.New("order already taken by another courier")
	ErrOrderNotAssigned  = errors.New("order not assigned to this courier")
	ErrOrderWrongStatus  = errors.New("order is in wrong status for this operation")
	ErrTooManyActive     = errors.New("courier has too many active orders")
)

type CourierService struct {
	pb.UnimplementedCourierAPIServer
	pgRepo  *repository.PostgresRepo
	metrics *metrics.Metrics
}

func NewCourierService(pgRepo *repository.PostgresRepo, m *metrics.Metrics) *CourierService {
	return &CourierService{
		pgRepo:  pgRepo,
		metrics: m,
	}
}

// Получение доступных заказов
type GetAvailableOrdersResult struct {
	Orders []*commonpb.Order
}

func (s *CourierService) GetAvailableOrders(ctx context.Context, courierID string, limit int32) (*GetAvailableOrdersResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	orders, err := s.pgRepo.ListAvailableOrders(ctx, limit, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list available orders: %v", err)
	}

	var pbOrders []*commonpb.Order
	for _, o := range orders {
		items, _ := s.pgRepo.GetOrderItems(ctx, o.ID)
		var pbItems []*commonpb.OrderItem
		for _, item := range items {
			pbItems = append(pbItems, &commonpb.OrderItem{
				ProductId: item.ProductID,
				Name:      item.Name,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
		}

		pbOrders = append(pbOrders, &commonpb.Order{
			Id:           o.ID,
			RestaurantId: o.RestaurantID,
			CourierId:    o.CourierID.String,
			ClientId:     o.UserID,
			Address:      o.Address,
			Items:        pbItems,
			TotalPrice:   o.TotalPrice,
			Status:       common_methods.MapOrderStatus(o.Status),
		})
	}

	return &GetAvailableOrdersResult{
		Orders: pbOrders,
	}, nil
}

// Принять заказ
type AcceptOrderResult struct {
	Success bool
	Order   *commonpb.Order
}

func (s *CourierService) AcceptOrder(ctx context.Context, courierID, orderID string) (*AcceptOrderResult, error) {
	// Проверяем, что у курьера нет слишком много активных заказов
	activeCount, err := s.pgRepo.CountActiveOrdersByCourier(ctx, courierID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count active orders: %v", err)
	}
	if activeCount >= 3 {
		return nil, status.Error(codes.FailedPrecondition, ErrTooManyActive.Error())
	}

	// Атомарно назначаем курьера и меняем статус на 'delivering'
	err = s.pgRepo.AssignCourierToOrder(ctx, orderID, courierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "order not available")
		}
		return nil, status.Errorf(codes.Internal, "assign courier: %v", err)
	}

	// Получаем полную информацию о заказе
	order, err := s.pgRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}

	items, _ := s.pgRepo.GetOrderItems(ctx, orderID)
	var pbItems []*commonpb.OrderItem
	for _, item := range items {
		pbItems = append(pbItems, &commonpb.OrderItem{
			ProductId: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	pbOrder := &commonpb.Order{
		Id:           order.ID,
		RestaurantId: order.RestaurantID,
		CourierId:    courierID,
		ClientId:     order.UserID,
		Address:      order.Address,
		Items:        pbItems,
		TotalPrice:   order.TotalPrice,
		Status:       commonpb.OrderStatus_ORDER_STATUS_DELIVERING,
	}

	// TODO: Отправить событие в Kafka — курьер принял заказ

	return &AcceptOrderResult{
		Success: true,
		Order:   pbOrder,
	}, nil
}

// Все заказы курьера
type GetMyOrdersResult struct {
	Orders []*commonpb.Order
}

func (s *CourierService) GetMyOrders(ctx context.Context, courierID string) (*GetMyOrdersResult, error) {
	orders, err := s.pgRepo.ListOrdersByCourier(ctx, courierID, 100, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list courier orders: %v", err)
	}

	var pbOrders []*commonpb.Order
	for _, o := range orders {
		items, _ := s.pgRepo.GetOrderItems(ctx, o.ID)
		var pbItems []*commonpb.OrderItem
		for _, item := range items {
			pbItems = append(pbItems, &commonpb.OrderItem{
				ProductId: item.ProductID,
				Name:      item.Name,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
		}

		pbOrders = append(pbOrders, &commonpb.Order{
			Id:           o.ID,
			RestaurantId: o.RestaurantID,
			CourierId:    o.CourierID.String,
			ClientId:     o.UserID,
			Address:      o.Address,
			Items:        pbItems,
			TotalPrice:   o.TotalPrice,
			Status:       common_methods.MapOrderStatus(o.Status),
		})
	}

	return &GetMyOrdersResult{
		Orders: pbOrders,
	}, nil
}

// Забрать заказ
type PickUpOrderResult struct {
	Success bool
	Status  commonpb.OrderStatus
}

func (s *CourierService) PickUpOrder(ctx context.Context, courierID, orderID string) (*PickUpOrderResult, error) {
	// Проверяем, что заказ назначен этому курьеру
	belongs, err := s.pgRepo.CheckOrderAssignedToCourier(ctx, orderID, courierID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check order assignment: %v", err)
	}
	if !belongs {
		return nil, status.Error(codes.PermissionDenied, ErrOrderNotAssigned.Error())
	}

	// Проверяем статус заказа
	order, err := s.pgRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}
	if order.Status != "delivering" {
		return nil, status.Error(codes.FailedPrecondition, "order must be in delivering status")
	}

	// Обновляем статус на 'delivering'
	if err := s.pgRepo.UpdateOrderStatus(ctx, orderID, "delivering"); err != nil {
		return nil, status.Errorf(codes.Internal, "update order status: %v", err)
	}

	// TODO: Отправить событие в Kafka — курьер забрал заказ

	s.metrics.OnOrderPickUp()

	return &PickUpOrderResult{
		Success: true,
		Status:  commonpb.OrderStatus_ORDER_STATUS_DELIVERING,
	}, nil
}

// Сдать заказ
type DeliverOrderResult struct {
	Success  bool
	Status   commonpb.OrderStatus
	Earnings int64
}

func (s *CourierService) DeliverOrder(ctx context.Context, courierID, orderID string) (*DeliverOrderResult, error) {
	// Проверяем, что заказ назначен этому курьеру
	belongs, err := s.pgRepo.CheckOrderAssignedToCourier(ctx, orderID, courierID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check order assignment: %v", err)
	}
	if !belongs {
		return nil, status.Error(codes.PermissionDenied, ErrOrderNotAssigned.Error())
	}

	// Проверяем статус заказа
	order, err := s.pgRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}
	if order.Status != "picked_up" {
		return nil, status.Error(codes.FailedPrecondition, "order must be picked up before delivery")
	}

	// Обновляем статус на 'delivered'
	if err := s.pgRepo.UpdateOrderStatus(ctx, orderID, "delivered"); err != nil {
		return nil, status.Errorf(codes.Internal, "update order status: %v", err)
	}

	s.metrics.OnOrderDelivered()

	earnings := order.TotalPrice * 15 / 100

	// TODO: Отправить событие в Kafka — заказ доставлен
	// TODO: Начислить деньги курьеру

	s.metrics.OnOrderDelivered()

	return &DeliverOrderResult{
		Success:  true,
		Status:   commonpb.OrderStatus_ORDER_STATUS_DELIVERED,
		Earnings: earnings,
	}, nil
}
