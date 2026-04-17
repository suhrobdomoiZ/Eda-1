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

func (s *Restaurant) UpdateProduct(ctx context.Context, product *models.FullProduct) (uuid.UUID, error) {
	productId, err := s.repo.UpdateProductInMenu(ctx, product)
	if err != nil {
		return uuid.Nil, err
	}

	return productId, nil
}

func (s *Restaurant) DeleteProduct(ctx context.Context, productId *models.ProductId) error {
	err := s.repo.DeleteProductFromMenu(ctx, productId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Restaurant) ListProducts(ctx context.Context, restaurantId *models.RestaurantId) ([]models.FullProduct, error) {
	result, err := s.repo.ListProducts(ctx, restaurantId)
	if err != nil {
		return nil, err
	}

	return result, err
}
