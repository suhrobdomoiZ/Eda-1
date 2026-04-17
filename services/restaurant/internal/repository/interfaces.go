package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
)

type repositoryCtxtKey string

const KeyTx repositoryCtxtKey = "pgx_tx"

type IExecutor interface {
	Exec(
		ctx context.Context,
		sql string,
		arguments ...any,
	) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type IRestaurant interface {
	AddProductIntoMenu(context.Context, *models.ProductInfo) (uuid.UUID, error)
	UpdateProductInMenu(context.Context, *models.FullProduct) (uuid.UUID, error)
	DeleteProductFromMenu(ctx context.Context, productId *models.ProductId) error
	ListProducts(ctx context.Context, restaurantId *models.RestaurantId) ([]models.FullProduct, error)
}
