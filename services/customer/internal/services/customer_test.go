// services/customer/internal/services/service_test.go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

// ============================================================================
// Mocks
// ============================================================================

type mockPgRepo struct {
	mock.Mock
}

func (m *mockPgRepo) CreateOrder(ctx context.Context, order *repository.Order, items []repository.OrderItem) error {
	args := m.Called(ctx, order, items)
	return args.Error(0)
}

func (m *mockPgRepo) GetOrderByID(ctx context.Context, orderID string) (*repository.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Order), args.Error(1)
}

func (m *mockPgRepo) GetOrderItems(ctx context.Context, orderID string) ([]repository.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.OrderItem), args.Error(1)
}

func (m *mockPgRepo) GetOrderWithItems(ctx context.Context, orderID string) (*repository.OrderWithItems, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.OrderWithItems), args.Error(1)
}

func (m *mockPgRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *mockPgRepo) CancelOrder(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *mockPgRepo) ListOrdersByUserID(ctx context.Context, userID string, limit, offset int32) ([]repository.OrderListItem, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.OrderListItem), args.Error(1)
}

func (m *mockPgRepo) CountOrdersByUserID(ctx context.Context, userID string) (int32, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockPgRepo) CheckOrderBelongsToUser(ctx context.Context, orderID, userID string) (bool, error) {
	args := m.Called(ctx, orderID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockPgRepo) Close() error {
	return nil
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

func (m *mockMetrics) OnOrderCreated(price float64) {
	m.Called(price)
}

func (m *mockMetrics) OnOrderCancelled() {
	m.Called()
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestCustomerService(pg *mockPgRepo, producer *mockProducer, metrics *mockMetrics) *service.CustomerService {
	return service.NewCustomerService(pg, producer, metrics)
}

func validCreateOrderInput() *service.CreateOrderInput {
	return &service.CreateOrderInput{
		UserID:       "user-123",
		RestaurantID: "rest-456",
		Items: []*pb.CreateOrderItem{
			{ProductId: "prod-1", Quantity: 2},
			{ProductId: "prod-2", Quantity: 1},
		},
		Address: "ул. Пушкина, д. 10",
	}
}

// ============================================================================
// CreateOrder Tests
// ============================================================================

func TestCreateOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	input := validCreateOrderInput()

	pg.On("CreateOrder", mock.Anything, mock.MatchedBy(func(o *repository.Order) bool {
		return o.UserID == input.UserID && o.RestaurantID == input.RestaurantID
	}), mock.Anything).Return(nil)

	producer.On("Send", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	metrics.On("OnOrderCreated", mock.AnythingOfType("float64")).Return()

	result, err := svc.CreateOrder(context.Background(), input)

	require.NoError(t, err)
	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "created", result.Status)
	pg.AssertExpectations(t)
}

func TestCreateOrder_Validation_EmptyRestaurant(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	metrics.On("IncError", "create_order", "validation").Return()

	input := validCreateOrderInput()
	input.RestaurantID = ""

	_, err := svc.CreateOrder(context.Background(), input)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "restaurant_id")
}

func TestCreateOrder_Validation_EmptyItems(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	metrics.On("IncError", "create_order", "validation").Return()

	input := validCreateOrderInput()
	input.Items = nil

	_, err := svc.CreateOrder(context.Background(), input)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "items")
}

func TestCreateOrder_Validation_EmptyAddress(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	metrics.On("IncError", "create_order", "validation").Return()

	input := validCreateOrderInput()
	input.Address = ""

	_, err := svc.CreateOrder(context.Background(), input)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "address")
}

func TestCreateOrder_DatabaseError(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	input := validCreateOrderInput()
	dbErr := errors.New("connection refused")

	pg.On("CreateOrder", mock.Anything, mock.Anything, mock.Anything).Return(dbErr)
	metrics.On("IncError", "create_order", "database").Return()

	_, err := svc.CreateOrder(context.Background(), input)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ============================================================================
// GetOrder Tests
// ============================================================================

func TestGetOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	orderID := uuid.New().String()
	userID := "user-123"

	pg.On("CheckOrderBelongsToUser", mock.Anything, orderID, userID).Return(true, nil)

	order := &repository.Order{
		ID:           orderID,
		UserID:       userID,
		RestaurantID: "rest-456",
		Address:      "ул. Пушкина",
		TotalPrice:   50000,
		Status:       "created",
	}
	items := []repository.OrderItem{
		{ID: "item-1", OrderID: orderID, ProductID: "prod-1", Name: "Пицца", Quantity: 2, Price: 25000},
	}

	pg.On("GetOrderWithItems", mock.Anything, orderID).Return(&repository.OrderWithItems{
		Order: *order,
		Items: items,
	}, nil)

	result, err := svc.GetOrder(context.Background(), userID, orderID)

	require.NoError(t, err)
	assert.Equal(t, orderID, result.Order.Id)
	assert.Equal(t, int64(50000), result.Order.TotalPrice)
	assert.Len(t, result.Order.Items, 1)
}

func TestGetOrder_NotBelongsToUser(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	orderID := uuid.New().String()
	userID := "user-123"

	pg.On("CheckOrderBelongsToUser", mock.Anything, orderID, userID).Return(false, nil)
	metrics.On("IncError", "get_order", "validation").Return()

	_, err := svc.GetOrder(context.Background(), userID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestGetOrder_NotFound(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	orderID := uuid.New().String()
	userID := "user-123"

	pg.On("CheckOrderBelongsToUser", mock.Anything, orderID, userID).Return(true, nil)
	pg.On("GetOrderWithItems", mock.Anything, orderID).Return(nil, repository.ErrNotFound)
	metrics.On("IncError", "get_order", "not_found").Return()

	_, err := svc.GetOrder(context.Background(), userID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

// ============================================================================
// CancelOrder Tests
// ============================================================================

func TestCancelOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	orderID := uuid.New().String()
	userID := "user-123"

	pg.On("CheckOrderBelongsToUser", mock.Anything, orderID, userID).Return(true, nil)

	order := &repository.Order{
		ID:         orderID,
		UserID:     userID,
		TotalPrice: 50000,
		Status:     "created",
	}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	pg.On("CancelOrder", mock.Anything, orderID).Return(nil)

	producer.On("Send", mock.Anything, orderID, mock.Anything).Return(nil)
	metrics.On("OnOrderCancelled").Return()

	result, err := svc.CancelOrder(context.Background(), userID, orderID)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(50000), result.RefundAmount)
}

func TestCancelOrder_CannotCancel_Cooking(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	orderID := uuid.New().String()
	userID := "user-123"

	pg.On("CheckOrderBelongsToUser", mock.Anything, orderID, userID).Return(true, nil)

	order := &repository.Order{
		ID:     orderID,
		UserID: userID,
		Status: "cooking",
	}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	metrics.On("IncError", "cancel_order", "validation").Return()

	_, err := svc.CancelOrder(context.Background(), userID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())

	// Проверить, что CancelOrder НЕ вызывался
	pg.AssertNotCalled(t, "CancelOrder", mock.Anything, orderID)
}

// ============================================================================
// ListMyOrders Tests
// ============================================================================

func TestListMyOrders_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	userID := "user-123"

	orders := []repository.OrderListItem{
		{ID: "order-1", RestaurantName: "Пицца Миа", Status: "delivered", TotalPrice: 50000},
		{ID: "order-2", RestaurantName: "Суши Мастер", Status: "created", TotalPrice: 30000},
	}

	pg.On("ListOrdersByUserID", mock.Anything, userID, int32(20), int32(0)).Return(orders, nil)
	pg.On("CountOrdersByUserID", mock.Anything, userID).Return(int32(2), nil)

	result, err := svc.ListMyOrders(context.Background(), userID, 20, 0)

	require.NoError(t, err)
	assert.Len(t, result.Orders, 2)
	assert.Equal(t, int32(2), result.Total)
}

func TestListMyOrders_DefaultLimit(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	userID := "user-123"

	pg.On("ListOrdersByUserID", mock.Anything, userID, int32(20), int32(0)).Return([]repository.OrderListItem{}, nil)
	pg.On("CountOrdersByUserID", mock.Anything, userID).Return(int32(0), nil)

	_, err := svc.ListMyOrders(context.Background(), userID, 0, 0)

	require.NoError(t, err)
	pg.AssertCalled(t, "ListOrdersByUserID", mock.Anything, userID, int32(20), int32(0))
}

func TestListMyOrders_MaxLimit(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	svc := newTestCustomerService(pg, producer, metrics)

	userID := "user-123"

	pg.On("ListOrdersByUserID", mock.Anything, userID, int32(100), int32(0)).Return([]repository.OrderListItem{}, nil)
	pg.On("CountOrdersByUserID", mock.Anything, userID).Return(int32(0), nil)

	_, err := svc.ListMyOrders(context.Background(), userID, 200, 0)

	require.NoError(t, err)
	pg.AssertCalled(t, "ListOrdersByUserID", mock.Anything, userID, int32(100), int32(0))
}
