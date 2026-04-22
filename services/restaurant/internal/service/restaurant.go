package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/repository"
)

type Producer interface {
	Send(ctx context.Context, key string, payload any) error
	Close() error
}

type Metrics interface {
	IncError(method, errorType string)
}

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type Restaurant struct {
	repo     repository.IRestaurant
	producer Producer
	logger   Logger
	metrics  Metrics
}

func NewRestaurant(
	repo repository.IRestaurant,
	producer Producer,
	metrics Metrics,
	logger Logger,
) *Restaurant {
	return &Restaurant{
		repo:     repo,
		producer: producer,
		metrics:  metrics,
		logger:   logger,
	}
}

func (s *Restaurant) AddProduct(ctx context.Context, productInfo *models.ProductInfo) (uuid.UUID, error) {

	productId, err := s.repo.AddProductIntoMenu(ctx, productInfo)
	if err != nil {
		s.metrics.IncError("add_product", metrics.ErrorTypeDatabase)
		slog.Error("AddProduct failed", "error", err)
		return uuid.Nil, err
	}

	slog.Info("AddProduct successfully done")

	return productId, nil
}

func (s *Restaurant) UpdateProduct(ctx context.Context, product *models.FullProduct) (uuid.UUID, error) {
	productId, err := s.repo.UpdateProductInMenu(ctx, product)
	if err != nil {
		s.metrics.IncError("update_product", metrics.ErrorTypeDatabase)
		slog.Error("UpdateProduct failed", "error", err)
		return uuid.Nil, err
	}

	slog.Info("UpdateProduct successfully done")

	return productId, nil
}

func (s *Restaurant) DeleteProduct(ctx context.Context, productId *models.ProductId) error {
	err := s.repo.DeleteProductFromMenu(ctx, productId)
	if err != nil {
		s.metrics.IncError("delete_product", metrics.ErrorTypeDatabase)
		return err
	}

	return nil
}

func (s *Restaurant) ListProducts(ctx context.Context, restaurantId *models.RestaurantId) ([]models.FullProduct, error) {
	result, err := s.repo.ListProducts(ctx, restaurantId)
	if err != nil {
		s.metrics.IncError("list_products", metrics.ErrorTypeDatabase)
		return nil, err
	}

	return result, err
}

func (s *Restaurant) GetProduct(ctx context.Context, productId *models.ProductId) (*models.FullProduct, error) {
	result, err := s.repo.GetProduct(ctx, productId)
	if err != nil {
		s.metrics.IncError("get_product", metrics.ErrorTypeDatabase)
		slog.Error("GetProduct failed", "error", err)
		return nil, err
	}

	s.logger.Info("GetProduct successfully done")

	return result, nil
}

func (s *Restaurant) ChangeOrderStatus(ctx context.Context, order *models.OrderIdWithStatus) (*models.ChangedOrderId, error) {

	resultId, err := s.repo.ChangeOrderStatus(ctx, order)
	if err != nil {
		s.metrics.IncError("change_order_status", metrics.ErrorTypeDatabase)
		s.logger.Error("ChangeOrderStatus failed", "error", err)
		return nil, err
	}
	s.logger.Info("ChangeOrderStatus successfully done")

	event := kafka.ChangeOrderStatusEvent{
		OrderId:   order.OrderId,
		NewStatus: order.Status,
	}
	err = s.producer.Send(ctx, order.OrderId.String(), event)
	if err != nil {
		s.metrics.IncError("add_product", metrics.ErrorTypeKafka)
		s.logger.Error("producer.Send failed", "error", err)
		return nil, err
	}

	s.logger.Info("producer.Send successfully done")

	return &models.ChangedOrderId{OrderId: resultId}, nil
}

func (s *Restaurant) ListOrders(ctx context.Context, restaurantId *models.RestaurantId) ([]models.Order, error) {
	result, err := s.repo.ListOrders(ctx, restaurantId)
	if err != nil {
		s.metrics.IncError("list_orders", metrics.ErrorTypeDatabase)
		s.logger.Error("ListOrders failed", "error", err)
		return nil, err
	}

	s.logger.Info("ListOrders successfully done")

	return result, err
}

func (s *Restaurant) ListRestaurants(ctx context.Context, limit, offset int32) ([]models.RestaurantInfo, int32, error) {
	restaurants, total, err := s.repo.ListRestaurants(ctx, limit, offset)
	if err != nil {
		s.metrics.IncError("list_orders", metrics.ErrorTypeDatabase)
		s.logger.Error("ListRestaurants failed", "error", err)
		return nil, 0, err
	}

	s.logger.Info("ListRestaurants successfully done")
	return restaurants, total, nil
}
