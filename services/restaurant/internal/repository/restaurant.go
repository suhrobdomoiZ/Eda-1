package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Restaurant struct {
	pool *pgxpool.Pool
}

func NewRestaurant(pool *pgxpool.Pool) *Restaurant {
	return &Restaurant{pool: pool}
}

//TODO:implement IRestaurant
