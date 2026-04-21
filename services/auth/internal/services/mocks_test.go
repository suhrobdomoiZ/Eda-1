package services_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/repository"
)

// Mock PostgresRepo

type mockPgRepo struct {
	mock.Mock
}

func (m *mockPgRepo) CreateUser(ctx context.Context, u *repository.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockPgRepo) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *mockPgRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *mockPgRepo) CreateRestaurantProfile(ctx context.Context, p *repository.RestaurantProfile) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPgRepo) GetRestaurantProfile(ctx context.Context, userID string) (*repository.RestaurantProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RestaurantProfile), args.Error(1)
}

func (m *mockPgRepo) CreateCourierProfile(ctx context.Context, p *repository.CourierProfile) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPgRepo) GetCourierProfile(ctx context.Context, userID string) (*repository.CourierProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.CourierProfile), args.Error(1)
}

// Mock RedisRepo

type mockRedisRepo struct {
	mock.Mock
}

func (m *mockRedisRepo) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return m.Called(ctx, userID, token, ttl).Error(0)
}

func (m *mockRedisRepo) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockRedisRepo) DeleteRefreshToken(ctx context.Context, userID, token string) error {
	return m.Called(ctx, userID, token).Error(0)
}

func (m *mockRedisRepo) DeleteAllUserTokens(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
