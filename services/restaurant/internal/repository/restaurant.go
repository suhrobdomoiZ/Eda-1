package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
)

type Restaurant struct {
	*Executor
}

func NewRestaurant(executor Executor) *Restaurant {
	return &Restaurant{&executor}
}

func (r *Restaurant) AddProductIntoMenu(ctx context.Context, productInfo *models.ProductInfo) (*uuid.UUID, error) {
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
		return nil, err
	}

	return &productId, nil
}

//TODO:implement IRestaurant
