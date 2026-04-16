package service

import (
	"database/sql"
)

type Restaurant struct {
}

func NewRestaurant(db *sql.DB) *Restaurant {
	return &Restaurant{db: db}
}
