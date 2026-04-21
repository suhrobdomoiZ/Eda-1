package repository

import (
	"context"
	"time"
)

type PgRepo interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateRestaurantProfile(ctx context.Context, p *RestaurantProfile) error
	GetRestaurantProfile(ctx context.Context, userID string) (*RestaurantProfile, error)
	CreateCourierProfile(ctx context.Context, p *CourierProfile) error
	GetCourierProfile(ctx context.Context, userID string) (*CourierProfile, error)
}

type RedisRepo interface {
	SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID, token string) error
	DeleteAllUserTokens(ctx context.Context, userID string) error
}
