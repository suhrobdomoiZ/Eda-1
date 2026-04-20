package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/service"
)

func NewOrderConsumerHandler(svc *service.Restaurant, logger *slog.Logger) func(ctx context.Context, key string, value []byte) error {
	return func(ctx context.Context, key string, value []byte) error {
		var evt kafka.ChangeOrderStatusEvent
		if err := json.Unmarshal(value, &evt); err != nil {
			return fmt.Errorf("unmarshal event: %w", err)
		}

		switch evt.NewStatus {
		case common.OrderStatus_ORDER_STATUS_CREATED:
			_, err := svc.ChangeOrderStatus(ctx, &models.OrderIdWithStatus{OrderId: evt.OrderId, Status: common.OrderStatus_ORDER_STATUS_COOKING})
			if err != nil {
				logger.Error("Failed to change status to ready", "err", err)
				return fmt.Errorf("OrderConsumerHandler: ChangeOrderStatus: %w", err)
			}

			time.Sleep(time.Second * 30)

			_, err = svc.ChangeOrderStatus(ctx, &models.OrderIdWithStatus{OrderId: evt.OrderId, Status: common.OrderStatus_ORDER_STATUS_READY})
			if err != nil {
				logger.Error("Failed to change status to ready", "err", err)
				return fmt.Errorf("OrderConsumerHandler: ChangeOrderStatus: %w", err)
			}

			logger.Info("Order is ready to be delivered", "order_id", evt.OrderId)
		}
		return nil
	}
}
