package closer

import (
	"context"
	"log/slog"
	"sync"

	"go.uber.org/multierr"
)

type closeFn func(ctx context.Context) error

type item struct {
	name string
	fn   closeFn
}

type Closer struct {
	logger slog.Logger
	mu     sync.Mutex
	items  []item
}

func New(logger slog.Logger) *Closer {
	return &Closer{
		logger: logger, //остальное по дефолту добавится
	}
}

func (c *Closer) Add(name string, fn func(ctx context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = append(c.items, item{name, fn})
}

func (c *Closer) AddFunc(name string, fn func()) {
	c.Add(name, func(ctx context.Context) error {
		fn()

		return nil
	})
}

func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	items := append([]item(nil), c.items...)
	c.mu.Unlock()

	var result error
	for _, item := range items {
		if err := item.fn(ctx); err != nil {
			result = multierr.Append(result, err)
			c.logger.Error("closer.Close: shutdown hook failed", "name", item.name, "err", err)
		}
	}
	c.logger.Info("closer.Close: shutdown hook finished", "result", result)

	return result
}
