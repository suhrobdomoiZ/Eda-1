package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
)

type ProductInfo struct {
	RestaurantId uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
}
