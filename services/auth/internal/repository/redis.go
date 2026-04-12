package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Префиксы ключей
const (
	prefixRefreshToken = "refresh:"     // refresh:TOKEN - user_id
	prefixUserTokens   = "user_tokens:" // user_tokens:USER_ID - set[TOKEN]
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(addr, password string, db int) (*RedisRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisRepo{client: client}, nil
}

// SaveRefreshToken сохраняет токен с привязкой к пользователю
func (r *RedisRepo) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	// token - userID (для валидации)
	pipe.Set(ctx, prefixRefreshToken+token, userID, ttl)
	// userID - set of tokens (для логаута со всех устройств)
	pipe.SAdd(ctx, prefixUserTokens+userID, token)
	pipe.Expire(ctx, prefixUserTokens+userID, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// GetUserIDByRefreshToken возвращает userID если токен валиден
func (r *RedisRepo) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := r.client.Get(ctx, prefixRefreshToken+token).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return userID, nil
}

// DeleteRefreshToken инвалидирует один конкретный токен (logout с одного устройства)
func (r *RedisRepo) DeleteRefreshToken(ctx context.Context, userID, token string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, prefixRefreshToken+token)
	pipe.SRem(ctx, prefixUserTokens+userID, token)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteAllUserTokens инвалидирует все токены пользователя (смена пароля и т.п.)
func (r *RedisRepo) DeleteAllUserTokens(ctx context.Context, userID string) error {
	tokens, err := r.client.SMembers(ctx, prefixUserTokens+userID).Result()
	if err != nil {
		return fmt.Errorf("get user tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}
	pipe := r.client.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, prefixRefreshToken+t)
	}
	pipe.Del(ctx, prefixUserTokens+userID)
	_, err = pipe.Exec(ctx)
	return err
}
