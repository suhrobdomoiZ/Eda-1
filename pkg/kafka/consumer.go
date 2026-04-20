package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type HandlerFunc func(ctx context.Context, key string, value []byte) error

type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

func NewConsumer(cfg Config, logger *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  cfg.Brokers,
			Topic:    cfg.Topic,
			GroupID:  cfg.ConsumerGroup,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		logger: logger,
	}
}

func (c *Consumer) Start(ctx context.Context, handler HandlerFunc) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Error("kafka: read failed, retrying...", "error", err)
				time.Sleep(1 * time.Millisecond)
				continue
			}

			if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
				c.logger.Warn("kafka: handler failed",
					"key", string(msg.Key),
					"error", err,
				)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
