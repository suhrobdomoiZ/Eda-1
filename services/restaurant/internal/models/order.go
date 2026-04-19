package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
)

type DBOrderStatus string

type OrderIdWithStatus struct {
	OrderId uuid.UUID          `json:"order_id"`
	Status  common.OrderStatus `json:"order_status"`
}

func IsValidOrderStatus(s common.OrderStatus) bool {
	switch s {
	case common.OrderStatus_ORDER_STATUS_CREATED,
		common.OrderStatus_ORDER_STATUS_COOKING,
		common.OrderStatus_ORDER_STATUS_READY,
		common.OrderStatus_ORDER_STATUS_DELIVERING,
		common.OrderStatus_ORDER_STATUS_DELIVERED,
		common.OrderStatus_ORDER_STATUS_CANCELED:
		return true
	default:
		return false
	}
}

type ChangeOrderStatusEvent struct {
	OrderId   uuid.UUID          `json:"order_id"`
	NewStatus common.OrderStatus `json:"new_order_status"`
}

type ChangedOrderId struct {
	OrderId uuid.UUID `json:"order_id"`
}

type OrderedProduct struct {
	ProductId uuid.UUID `json:"product_id"`
	Name      string    `json:"product_name"`
	OrderId   uuid.UUID `json:"order_id"`
	Price     int64     `json:"price"`
	Quantity  int32     `json:"quantity"`
}
type Order struct {
	Id           uuid.UUID          `json:"order_id"`
	RestaurantId uuid.UUID          `json:"restaurant_id"`
	CourierId    uuid.UUID          `json:"courier_id"`
	ClientId     uuid.UUID          `json:"client_id"`
	Address      string             `json:"address"`
	OrderedItems []OrderedProduct   `json:"ordered_items"`
	TotalPrice   int64              `json:"total_price"`
	OrderStatus  common.OrderStatus `json:"order_status"`
}
