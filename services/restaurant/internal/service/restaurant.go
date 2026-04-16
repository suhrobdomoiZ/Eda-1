package service

import (
	"database/sql"
	"sync"

	"github.com/suhrobdomoiZ/Eda-1/services/api"
)

type Restaurant struct {
	api.RestaurantServer
	mu sync.Mutex
	db *sql.DB
}

func NewRestaurant(db *sql.DB) *Restaurant {
	return &Restaurant{db: db}
}
