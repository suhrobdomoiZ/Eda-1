package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

type Restaurant struct {
	Executor
}

func NewRestaurant(executor Executor) *Restaurant {
	return &Restaurant{executor}
}

func (r *Restaurant) AddProductIntoMenu(ctx context.Context, productInfo *models.ProductInfo) (uuid.UUID, error) {
	query := `
		INSERT INTO products(id, restaurant_id, name, description, price)
		VALUES($1, $2, $3, $4, $5);
    `
	productId := uuid.New()
	_, err := r.GetExecutor(ctx).Exec(
		ctx,
		query,
		productId,
		productInfo.RestaurantId,
		productInfo.Name,
		productInfo.Description,
		productInfo.Price,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return uuid.Nil, utils.ErrProductAlreadyExists
			case "23503":
				return uuid.Nil, utils.ErrInvalidRestaurantID
			case "23514":
				return uuid.Nil, utils.ErrValidationFailed
			}
		}
		return uuid.Nil, fmt.Errorf("add product: %w", err)
	}

	return productId, nil
}

func (r *Restaurant) UpdateProductInMenu(ctx context.Context, product *models.FullProduct) (uuid.UUID, error) {
	query := `
	UPDATE products
	SET
    name = $2,
    description = $3,
    price = $4,
    WHERE id = $1 AND restaurant_id = $5;
	`
	_, err := r.GetExecutor(ctx).Exec(
		ctx,
		query,
		product.Id,
		product.Name,
		product.Description,
		product.Price,
		product.RestaurantId,
	)

	//TODO: самостоятельно написать коды
	if err != nil {
		var pgErr *pgconn.PgError
		switch pgErr.Code {
		case "23503":
			return uuid.Nil, utils.ErrInvalidRestaurantID
		case "23514":
			return uuid.Nil, utils.ErrValidationFailed
		case "23502":
			return uuid.Nil, utils.ErrValidationFailed
		}
	}

	return product.Id, nil
}
