package models

import (
	"github.com/google/uuid"
)

type ProductInfo struct {
	RestaurantId uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
}

type FullProduct struct {
	Id           uuid.UUID `json:"product_id"`
	RestaurantId uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
}

type ProductId struct {
	Id uuid.UUID `json:"product_id"`
}

type RestaurantId struct {
	Id uuid.UUID `json:"restaurant_id"`
}
