package kafka

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
)

// Это то, что кладем в кафку
type ChangeOrderStatusEvent struct {
	OrderId   uuid.UUID          `json:"order_id"`
	NewStatus common.OrderStatus `json:"new_order_status"`
}
