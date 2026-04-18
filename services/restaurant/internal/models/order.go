package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

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

func NewChangeOrderStatusEvent(orderWithStatus *OrderIdWithStatus) *ChangeOrderStatusEvent {
	return &ChangeOrderStatusEvent{
		OrderId:   orderWithStatus.OrderId,
		NewStatus: orderWithStatus.Status,
	}
}

type ChangedOrderId struct {
	OrderId uuid.UUID `json:"order_id"`
}

func ConvertChangedOrderIdToChangeOrderStatusResponse(changedOrderId *ChangedOrderId) *api.ChangeOrderStatusResponse{
	return &api.ChangeOrderStatusResponse{
		Id: changedOrderId.OrderId.String(),
		Status: utils.StatusOK,

	}
}