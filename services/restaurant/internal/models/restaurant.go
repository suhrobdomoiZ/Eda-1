package models

import (
	"github.com/google/uuid"
)

type RestaurantInfo struct {
	ID      uuid.UUID
	Name    string
	Address string
	Phone   string
	Cuisine string
	Rating  int32
	IsOpen  bool
}
