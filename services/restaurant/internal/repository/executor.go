package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor struct {
	pool *pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) *Executor {
	return &Executor{pool: pool}
}

func (e *Executor) GetExecutor(ctx context.Context) IExecutor {
	tx, ok := ctx.Value(KeyTx).(pgx.Tx)
	if ok {
		return tx
	}

	return e.pool
}
