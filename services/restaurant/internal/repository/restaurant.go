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
		// 👇 Проверяем, является ли ошибка специфичной ошибкой PostgreSQL
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": //TODO:найти коды ошибок, наподобие http.StatusOK
				return uuid.Nil, utils.ErrProductAlreadyExists
			case "23503": // foreign_key_violation
				return uuid.Nil, utils.ErrInvalidRestaurantID
			case "23514": // check_violation
				return uuid.Nil, utils.ErrValidationFailed
			}
		}
		// 👇 Всё остальное — оборачиваем с контекстом
		return uuid.Nil, fmt.Errorf("add product: %w", err)
	}

	return productId, nil
}

//TODO:implement IRestaurant
