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

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/courier/internal/services"
)

// ============================================================================
// Mocks
// ============================================================================

type mockPgRepo struct {
	mock.Mock
}

func (m *mockPgRepo) GetOrderByID(ctx context.Context, orderID string) (*repository.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Order), args.Error(1)
}

func (m *mockPgRepo) ListAvailableOrders(ctx context.Context, limit, offset int32) ([]repository.Order, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.Order), args.Error(1)
}

func (m *mockPgRepo) CountAvailableOrders(ctx context.Context) (int32, error) {
	args := m.Called(ctx)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockPgRepo) AssignCourierToOrder(ctx context.Context, orderID, courierID string) error {
	args := m.Called(ctx, orderID, courierID)
	return args.Error(0)
}

func (m *mockPgRepo) CheckOrderAssignedToCourier(ctx context.Context, orderID, courierID string) (bool, error) {
	args := m.Called(ctx, orderID, courierID)
	return args.Bool(0), args.Error(1)
}

func (m *mockPgRepo) ListOrdersByCourier(ctx context.Context, courierID string, limit, offset int32) ([]repository.Order, error) {
	args := m.Called(ctx, courierID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.Order), args.Error(1)
}

func (m *mockPgRepo) CountOrdersByCourier(ctx context.Context, courierID string) (int32, error) {
	args := m.Called(ctx, courierID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockPgRepo) CountActiveOrdersByCourier(ctx context.Context, courierID string) (int32, error) {
	args := m.Called(ctx, courierID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockPgRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *mockPgRepo) GetOrderItems(ctx context.Context, orderID string) ([]repository.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.OrderItem), args.Error(1)
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

func (m *mockMetrics) OnOrderPickUp() {
	m.Called()
}

func (m *mockMetrics) OnOrderDelivered() {
	m.Called()
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestCourierService(pg *mockPgRepo, producer *mockProducer, metrics *mockMetrics, logger *mockLogger) *service.CourierService {
	return service.NewCourierService(pg, producer, metrics, logger)
}

// ============================================================================
// GetAvailableOrders Tests
// ============================================================================

func TestGetAvailableOrders_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orders := []repository.Order{
		{ID: "order-1", RestaurantID: "rest-1", UserID: "user-1", Address: "ул. Ленина 1", TotalPrice: 50000, Status: "ready"},
		{ID: "order-2", RestaurantID: "rest-2", UserID: "user-2", Address: "ул. Мира 5", TotalPrice: 30000, Status: "ready"},
	}

	pg.On("ListAvailableOrders", mock.Anything, int32(20), int32(0)).Return(orders, nil)
	pg.On("GetOrderItems", mock.Anything, "order-1").Return([]repository.OrderItem{}, nil)
	pg.On("GetOrderItems", mock.Anything, "order-2").Return([]repository.OrderItem{}, nil)
	metrics.On("IncError", mock.Anything, mock.Anything).Maybe()

	result, err := svc.GetAvailableOrders(context.Background(), courierID, 20)

	require.NoError(t, err)
	assert.Len(t, result.Orders, 2)
	assert.Equal(t, "order-1", result.Orders[0].Id)
	assert.Equal(t, "order-2", result.Orders[1].Id)
	pg.AssertExpectations(t)
}

func TestGetAvailableOrders_DefaultLimit(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"

	pg.On("ListAvailableOrders", mock.Anything, int32(20), int32(0)).Return([]repository.Order{}, nil)
	pg.On("GetOrderItems", mock.Anything, mock.Anything).Maybe()
	metrics.On("IncError", mock.Anything, mock.Anything).Maybe()

	_, err := svc.GetAvailableOrders(context.Background(), courierID, 0)

	require.NoError(t, err)
	pg.AssertCalled(t, "ListAvailableOrders", mock.Anything, int32(20), int32(0))
}

func TestGetAvailableOrders_MaxLimit(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"

	pg.On("ListAvailableOrders", mock.Anything, int32(100), int32(0)).Return([]repository.Order{}, nil)
	pg.On("GetOrderItems", mock.Anything, mock.Anything).Maybe()
	metrics.On("IncError", mock.Anything, mock.Anything).Maybe()

	_, err := svc.GetAvailableOrders(context.Background(), courierID, 200)

	require.NoError(t, err)
	pg.AssertCalled(t, "ListAvailableOrders", mock.Anything, int32(100), int32(0))
}

func TestGetAvailableOrders_DatabaseError(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	dbErr := errors.New("connection refused")

	pg.On("ListAvailableOrders", mock.Anything, int32(20), int32(0)).Return(nil, dbErr)
	metrics.On("IncError", "list_available_orders", "database").Return()

	_, err := svc.GetAvailableOrders(context.Background(), courierID, 20)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	metrics.AssertCalled(t, "IncError", "list_available_orders", "database")
}

// ============================================================================
// AcceptOrder Tests
// ============================================================================

func TestAcceptOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CountActiveOrdersByCourier", mock.Anything, courierID).Return(int32(0), nil)
	pg.On("AssignCourierToOrder", mock.Anything, orderID, courierID).Return(nil)

	order := &repository.Order{
		ID:           orderID,
		RestaurantID: "rest-456",
		UserID:       "user-789",
		Address:      "ул. Пушкина",
		TotalPrice:   50000,
		Status:       "delivering",
	}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	pg.On("GetOrderItems", mock.Anything, orderID).Return([]repository.OrderItem{}, nil)

	producer.On("Send", mock.Anything, orderID, mock.Anything).Return(nil)
	logger.On("Info", "AcceptOrder success", mock.Anything).Return()

	result, err := svc.AcceptOrder(context.Background(), courierID, orderID)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Order)
	assert.Equal(t, orderID, result.Order.Id)
	assert.Equal(t, commonpb.OrderStatus_ORDER_STATUS_DELIVERING, result.Order.Status)
	producer.AssertCalled(t, "Send", mock.Anything, orderID, mock.Anything)
}

func TestAcceptOrder_TooManyActive(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CountActiveOrdersByCourier", mock.Anything, courierID).Return(int32(3), nil)

	_, err := svc.AcceptOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Contains(t, st.Message(), "too many active orders")
}

func TestAcceptOrder_OrderNotAvailable(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CountActiveOrdersByCourier", mock.Anything, courierID).Return(int32(0), nil)
	pg.On("AssignCourierToOrder", mock.Anything, orderID, courierID).Return(repository.ErrNotFound)
	metrics.On("IncError", "accept_order", "database").Return()

	_, err := svc.AcceptOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

// ============================================================================
// GetMyOrders Tests
// ============================================================================

func TestGetMyOrders_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"

	orders := []repository.Order{
		{ID: "order-1", RestaurantID: "rest-1", UserID: "user-1", TotalPrice: 50000, Status: "delivered"},
		{ID: "order-2", RestaurantID: "rest-2", UserID: "user-2", TotalPrice: 30000, Status: "delivering"},
	}

	pg.On("ListOrdersByCourier", mock.Anything, courierID, int32(100), int32(0)).Return(orders, nil)
	pg.On("GetOrderItems", mock.Anything, "order-1").Return([]repository.OrderItem{}, nil)
	pg.On("GetOrderItems", mock.Anything, "order-2").Return([]repository.OrderItem{}, nil)
	metrics.On("IncError", mock.Anything, mock.Anything).Maybe()

	result, err := svc.GetMyOrders(context.Background(), courierID)

	require.NoError(t, err)
	assert.Len(t, result.Orders, 2)
}

func TestGetMyOrders_DatabaseError(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	dbErr := errors.New("connection refused")

	pg.On("ListOrdersByCourier", mock.Anything, courierID, int32(100), int32(0)).Return(nil, dbErr)
	metrics.On("IncError", "list_orders_by_courier", "database").Return()

	_, err := svc.GetMyOrders(context.Background(), courierID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ============================================================================
// PickUpOrder Tests
// ============================================================================

func TestPickUpOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(true, nil)

	order := &repository.Order{ID: orderID, Status: "delivering"}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	pg.On("UpdateOrderStatus", mock.Anything, orderID, "delivering").Return(nil)

	metrics.On("OnOrderPickUp").Return()

	result, err := svc.PickUpOrder(context.Background(), courierID, orderID)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, commonpb.OrderStatus_ORDER_STATUS_DELIVERING, result.Status)
	metrics.AssertCalled(t, "OnOrderPickUp")
}

func TestPickUpOrder_NotAssigned(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(false, nil)
	metrics.On("IncError", "pick_up_order", "validation").Return()

	_, err := svc.PickUpOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestPickUpOrder_WrongStatus(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(true, nil)

	order := &repository.Order{ID: orderID, Status: "ready"} // не delivering
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)

	_, err := svc.PickUpOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Contains(t, st.Message(), "delivering status")
}

// ============================================================================
// DeliverOrder Tests
// ============================================================================

func TestDeliverOrder_Success(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(true, nil)

	order := &repository.Order{ID: orderID, Status: "delivering", TotalPrice: 100000}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	pg.On("UpdateOrderStatus", mock.Anything, orderID, "delivered").Return(nil)

	producer.On("Send", mock.Anything, orderID, mock.Anything).Return(nil)
	metrics.On("OnOrderDelivered").Return()

	result, err := svc.DeliverOrder(context.Background(), courierID, orderID)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, commonpb.OrderStatus_ORDER_STATUS_DELIVERED, result.Status)
	assert.Equal(t, int64(15000), result.Earnings) // 15% от 100000
	metrics.AssertCalled(t, "OnOrderDelivered")
	producer.AssertCalled(t, "Send", mock.Anything, orderID, mock.Anything)
}

func TestDeliverOrder_NotAssigned(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(false, nil)
	metrics.On("IncError", "check_order_assigned_to_courier", "validation").Return()

	_, err := svc.DeliverOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestDeliverOrder_WrongStatus(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(true, nil)

	order := &repository.Order{ID: orderID, Status: "ready"} // не delivering
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	metrics.On("IncError", "check_order_assigned_to_courier", "database").Return()

	_, err := svc.DeliverOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Contains(t, st.Message(), "delivering before delivery")
}

func TestDeliverOrder_DatabaseError(t *testing.T) {
	pg := &mockPgRepo{}
	producer := &mockProducer{}
	metrics := &mockMetrics{}
	logger := &mockLogger{}
	svc := newTestCourierService(pg, producer, metrics, logger)

	courierID := "courier-123"
	orderID := uuid.New().String()
	dbErr := errors.New("connection refused")

	pg.On("CheckOrderAssignedToCourier", mock.Anything, orderID, courierID).Return(true, nil)

	order := &repository.Order{ID: orderID, Status: "delivering", TotalPrice: 100000}
	pg.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	pg.On("UpdateOrderStatus", mock.Anything, orderID, "delivered").Return(dbErr)
	metrics.On("IncError", "check_order_assigned_to_courier", "database").Return()

	_, err := svc.DeliverOrder(context.Background(), courierID, orderID)

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
