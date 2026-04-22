package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
)

// ============================================================================
// Mocks
// ============================================================================

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) AddProductIntoMenu(ctx context.Context, productInfo *models.ProductInfo) (uuid.UUID, error) {
	args := m.Called(ctx, productInfo)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRepo) UpdateProductInMenu(ctx context.Context, product *models.FullProduct) (uuid.UUID, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRepo) DeleteProductFromMenu(ctx context.Context, productId *models.ProductId) error {
	args := m.Called(ctx, productId)
	return args.Error(0)
}

func (m *mockRepo) ListProducts(ctx context.Context, restaurantId *models.RestaurantId) ([]models.FullProduct, error) {
	args := m.Called(ctx, restaurantId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FullProduct), args.Error(1)
}

func (m *mockRepo) GetProduct(ctx context.Context, productId *models.ProductId) (*models.FullProduct, error) {
	args := m.Called(ctx, productId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FullProduct), args.Error(1)
}

func (m *mockRepo) ChangeOrderStatus(ctx context.Context, order *models.OrderIdWithStatus) (uuid.UUID, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRepo) ListOrders(ctx context.Context, restaurantId *models.RestaurantId) ([]models.Order, error) {
	args := m.Called(ctx, restaurantId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *mockRepo) ListRestaurants(ctx context.Context, limit, offset int32) ([]models.RestaurantInfo, int32, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.RestaurantInfo), args.Get(1).(int32), args.Error(2)
}

type mockProducer struct {
	mock.Mock
}

func (m *mockProducer) Send(ctx context.Context, key string, payload any) error {
	args := m.Called(ctx, key, payload)
	return args.Error(0)
}

func (m *mockProducer) Close() error {
	return nil
}

type mockMetrics struct {
	mock.Mock
}

func (m *mockMetrics) IncError(method, errorType string) {
	m.Called(method, errorType)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...any) {
	if len(args) == 0 {
		m.Called(msg)
	} else {
		callArgs := []interface{}{msg}
		callArgs = append(callArgs, args...)
		m.Called(callArgs...)
	}
}

func (m *mockLogger) Error(msg string, args ...any) {
	if len(args) == 0 {
		m.Called(msg)
	} else {
		callArgs := []interface{}{msg}
		callArgs = append(callArgs, args...)
		m.Called(callArgs...)
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestRestaurantService(repo *mockRepo, producer *mockProducer, metrics *mockMetrics, logger *mockLogger) *Restaurant {
	return NewRestaurant(repo, producer, metrics, logger)
}

func validProductInfo() *models.ProductInfo {
	return &models.ProductInfo{
		RestaurantId: uuid.New(),
		Name:         "Маргарита",
		Description:  "Томат, моцарелла, базилик",
		Price:        59000,
	}
}

func validFullProduct() *models.FullProduct {
	return &models.FullProduct{
		Id:           uuid.New(),
		RestaurantId: uuid.New(),
		Name:         "Маргарита",
		Description:  "Томат, моцарелла, базилик",
		Price:        59000,
	}
}

// ============================================================================
// AddProduct Tests
// ============================================================================

func TestAddProduct_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productInfo := validProductInfo()
	expectedID := uuid.New()

	repo.On("AddProductIntoMenu", mock.Anything, productInfo).Return(expectedID, nil)
	logger.On("Info", "AddProduct successfully done").Return()

	result, err := svc.AddProduct(context.Background(), productInfo)

	require.NoError(t, err)
	assert.Equal(t, expectedID, result)
	repo.AssertExpectations(t)
}

func TestAddProduct_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productInfo := validProductInfo()
	dbErr := errors.New("connection refused")

	repo.On("AddProductIntoMenu", mock.Anything, productInfo).Return(uuid.Nil, dbErr)
	metrics.On("IncError", "add_product", "database").Return()
	logger.On("Error", "AddProduct failed", "error", dbErr).Return()

	_, err := svc.AddProduct(context.Background(), productInfo)

	require.Error(t, err)
	assert.Equal(t, dbErr, err)
	metrics.AssertCalled(t, "IncError", "add_product", "database")
}

// ============================================================================
// UpdateProduct Tests
// ============================================================================

func TestUpdateProduct_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	product := validFullProduct()
	expectedID := product.Id

	repo.On("UpdateProductInMenu", mock.Anything, product).Return(expectedID, nil)
	logger.On("Info", "UpdateProduct successfully done").Return()

	result, err := svc.UpdateProduct(context.Background(), product)

	require.NoError(t, err)
	assert.Equal(t, expectedID, result)
}

func TestUpdateProduct_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	product := validFullProduct()
	dbErr := errors.New("connection refused")

	repo.On("UpdateProductInMenu", mock.Anything, product).Return(uuid.Nil, dbErr)
	metrics.On("IncError", "update_product", "database").Return()
	logger.On("Error", "UpdateProduct failed", "error", dbErr).Return()

	_, err := svc.UpdateProduct(context.Background(), product)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "update_product", "database")
}

// ============================================================================
// DeleteProduct Tests
// ============================================================================

func TestDeleteProduct_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productID := &models.ProductId{Id: uuid.New()}

	repo.On("DeleteProductFromMenu", mock.Anything, productID).Return(nil)

	err := svc.DeleteProduct(context.Background(), productID)

	require.NoError(t, err)
}

func TestDeleteProduct_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productID := &models.ProductId{Id: uuid.New()}
	dbErr := errors.New("connection refused")

	repo.On("DeleteProductFromMenu", mock.Anything, productID).Return(dbErr)
	metrics.On("IncError", "delete_product", "database").Return()

	err := svc.DeleteProduct(context.Background(), productID)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "delete_product", "database")
}

// ============================================================================
// ListProducts Tests
// ============================================================================

func TestListProducts_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	restaurantID := &models.RestaurantId{Id: uuid.New()}
	expectedProducts := []models.FullProduct{
		{Id: uuid.New(), Name: "Маргарита", Price: 59000},
		{Id: uuid.New(), Name: "Пепперони", Price: 69000},
	}

	repo.On("ListProducts", mock.Anything, restaurantID).Return(expectedProducts, nil)

	result, err := svc.ListProducts(context.Background(), restaurantID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Маргарита", result[0].Name)
}

func TestListProducts_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	restaurantID := &models.RestaurantId{Id: uuid.New()}
	dbErr := errors.New("connection refused")

	repo.On("ListProducts", mock.Anything, restaurantID).Return(nil, dbErr)
	metrics.On("IncError", "list_products", "database").Return()

	_, err := svc.ListProducts(context.Background(), restaurantID)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "list_products", "database")
}

// ============================================================================
// GetProduct Tests
// ============================================================================

func TestGetProduct_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productID := &models.ProductId{Id: uuid.New()}
	expectedProduct := &models.FullProduct{
		Id:    productID.Id,
		Name:  "Маргарита",
		Price: 59000,
	}

	repo.On("GetProduct", mock.Anything, productID).Return(expectedProduct, nil)
	logger.On("Info", "GetProduct successfully done").Return() // ← уже правильно

	result, err := svc.GetProduct(context.Background(), productID)

	require.NoError(t, err)
	assert.Equal(t, expectedProduct.Id, result.Id)
	assert.Equal(t, expectedProduct.Name, result.Name)
}

func TestGetProduct_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	productID := &models.ProductId{Id: uuid.New()}
	dbErr := errors.New("not found")

	repo.On("GetProduct", mock.Anything, productID).Return(nil, dbErr)
	metrics.On("IncError", "get_product", "database").Return()
	logger.On("Error", "GetProduct failed", "error", dbErr).Return()

	_, err := svc.GetProduct(context.Background(), productID)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "get_product", "database")
}

// ============================================================================
// ChangeOrderStatus Tests
// ============================================================================

func TestChangeOrderStatus_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	orderID := uuid.New()
	order := &models.OrderIdWithStatus{
		OrderId: orderID,
		Status:  commonpb.OrderStatus_ORDER_STATUS_COOKING,
	}

	repo.On("ChangeOrderStatus", mock.Anything, order).Return(orderID, nil)
	producer.On("Send", mock.Anything, orderID.String(), mock.Anything).Return(nil)
	logger.On("Info", "ChangeOrderStatus successfully done").Return()
	logger.On("Info", "producer.Send successfully done").Return()

	result, err := svc.ChangeOrderStatus(context.Background(), order)

	require.NoError(t, err)
	assert.Equal(t, orderID, result.OrderId)
	producer.AssertCalled(t, "Send", mock.Anything, orderID.String(), mock.Anything)
}

func TestChangeOrderStatus_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	order := &models.OrderIdWithStatus{
		OrderId: uuid.New(),
		Status:  commonpb.OrderStatus_ORDER_STATUS_COOKING,
	}
	dbErr := errors.New("connection refused")

	repo.On("ChangeOrderStatus", mock.Anything, order).Return(uuid.Nil, dbErr)
	metrics.On("IncError", "change_order_status", "database").Return()
	logger.On("Error", "ChangeOrderStatus failed", "error", dbErr).Return()

	_, err := svc.ChangeOrderStatus(context.Background(), order)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "change_order_status", "database")
}

func TestChangeOrderStatus_KafkaError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	orderID := uuid.New()
	order := &models.OrderIdWithStatus{
		OrderId: orderID,
		Status:  commonpb.OrderStatus_ORDER_STATUS_COOKING,
	}
	kafkaErr := errors.New("kafka unavailable")

	repo.On("ChangeOrderStatus", mock.Anything, order).Return(orderID, nil)
	producer.On("Send", mock.Anything, orderID.String(), mock.Anything).Return(kafkaErr)
	metrics.On("IncError", "add_product", "kafka").Return()
	logger.On("Info", "ChangeOrderStatus successfully done").Return()
	logger.On("Error", "producer.Send failed", "error", kafkaErr).Return()

	_, err := svc.ChangeOrderStatus(context.Background(), order)

	require.Error(t, err)
	assert.Equal(t, kafkaErr, err)
	metrics.AssertCalled(t, "IncError", "add_product", "kafka")
}

// ============================================================================
// ListOrders Tests
// ============================================================================

func TestListOrders_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	restaurantID := &models.RestaurantId{Id: uuid.New()}
	expectedOrders := []models.Order{
		{Id: uuid.New(), OrderStatus: commonpb.OrderStatus_ORDER_STATUS_CREATED, TotalPrice: 50000},
		{Id: uuid.New(), OrderStatus: commonpb.OrderStatus_ORDER_STATUS_COOKING, TotalPrice: 30000},
	}

	repo.On("ListOrders", mock.Anything, restaurantID).Return(expectedOrders, nil)
	logger.On("Info", "ListOrders successfully done").Return()

	result, err := svc.ListOrders(context.Background(), restaurantID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListOrders_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	restaurantID := &models.RestaurantId{Id: uuid.New()}
	dbErr := errors.New("connection refused")

	repo.On("ListOrders", mock.Anything, restaurantID).Return(nil, dbErr)
	metrics.On("IncError", "list_orders", "database").Return()
	logger.On("Error", "ListOrders failed", "error", dbErr).Return()

	_, err := svc.ListOrders(context.Background(), restaurantID)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "list_orders", "database")
}

// ============================================================================
// ListRestaurants Tests
// ============================================================================

func TestListRestaurants_Success(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	expectedRestaurants := []models.RestaurantInfo{
		{ID: uuid.New(), Name: "Пицца Миа", Cuisine: "Итальянская"},
		{ID: uuid.New(), Name: "Суши Мастер", Cuisine: "Японская"},
	}

	repo.On("ListRestaurants", mock.Anything, int32(20), int32(0)).Return(expectedRestaurants, int32(2), nil)
	logger.On("Info", "ListRestaurants successfully done").Return()

	result, total, err := svc.ListRestaurants(context.Background(), 20, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int32(2), total)
}

func TestListRestaurants_DatabaseError(t *testing.T) {
	repo := &mockRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestRestaurantService(repo, producer, metrics, logger)

	dbErr := errors.New("connection refused")

	repo.On("ListRestaurants", mock.Anything, int32(20), int32(0)).Return(nil, int32(0), dbErr)
	metrics.On("IncError", "list_orders", "database").Return()
	logger.On("Error", "ListRestaurants failed", "error", dbErr).Return()

	_, _, err := svc.ListRestaurants(context.Background(), 20, 0)

	require.Error(t, err)
	metrics.AssertCalled(t, "IncError", "list_orders", "database")
}
