package common

import (
	"context"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
)

// Название ресторана по ID
func GetRestaurantName(ctx context.Context, restaurantID string) string {
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

// str -> enum
func MapOrderStatus(status string) commonpb.OrderStatus {
	switch status {
	case "created":
		return commonpb.OrderStatus_ORDER_STATUS_CREATED
	case "cooking":
		return commonpb.OrderStatus_ORDER_STATUS_COOKING
	case "ready":
		return commonpb.OrderStatus_ORDER_STATUS_READY
	case "delivering":
		return commonpb.OrderStatus_ORDER_STATUS_DELIVERING
	case "delivered":
		return commonpb.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return commonpb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return commonpb.OrderStatus_ORDER_STATUS_UNKNOWN
	}
}

// enum -> str
func MapOrderStatusToString(status commonpb.OrderStatus) string {
	switch status {
	case commonpb.OrderStatus_ORDER_STATUS_CREATED:
		return "created"
	case commonpb.OrderStatus_ORDER_STATUS_COOKING:
		return "cooking"
	case commonpb.OrderStatus_ORDER_STATUS_READY:
		return "ready"
	case commonpb.OrderStatus_ORDER_STATUS_DELIVERING:
		return "delivering"
	case commonpb.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case commonpb.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unknown"
	}
}
