package service

import (
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/repository"
)

type Restaurant struct {
	repo repository.IRestaurant
}

func NewRestaurant(repository repository.IRestaurant) *Restaurant {
	return &Restaurant{repo: repository}
}

func AddProduct()
