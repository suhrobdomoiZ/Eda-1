package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/repository"
)

type Restaurant struct {
	repo repository.IRestaurant
}

func NewRestaurant(repository repository.IRestaurant) *Restaurant {
	return &Restaurant{repo: repository}
}

func (s *Restaurant) AddProduct(ctx context.Context, productInfo *models.ProductInfo) (uuid.UUID, error) {

	productId, err := s.repo.AddProductIntoMenu(ctx, productInfo)
	if err != nil {
		return uuid.Nil, err
	}

	return productId, nil
}
