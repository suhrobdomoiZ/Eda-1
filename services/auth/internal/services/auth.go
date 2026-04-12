package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	pg  *repository.PostgresRepo
	rdb *repository.RedisRepo
	jwt *JWTService
}

func NewAuthService(
	pg *repository.PostgresRepo,
	rdb *repository.RedisRepo,
	jwt *JWTService,
) *AuthService {
	return &AuthService{pg: pg, rdb: rdb, jwt: jwt}
}

// RegisterResult - то что возвращаем после успешной регистрации
type RegisterResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type RegisterInput struct {
	Username string
	Password string
	Role     string // "user", "restaurant", "courier"

	// Опциональные профили
	RestaurantName    string
	RestaurantAddress string
	RestaurantPhone   string

	CourierName  string
	CourierPhone string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.New().String()

	user := &repository.User{
		ID:           userID,
		Username:     in.Username,
		PasswordHash: string(hash),
		Role:         in.Role,
	}

	if err := s.pg.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Создаём профиль в зависимости от роли
	switch in.Role {
	case "restaurant":
		if err := s.pg.CreateRestaurantProfile(ctx, &repository.RestaurantProfile{
			UserID:  userID,
			Name:    in.RestaurantName,
			Address: in.RestaurantAddress,
			Phone:   in.RestaurantPhone,
		}); err != nil {
			return nil, fmt.Errorf("create restaurant profile: %w", err)
		}
	case "courier":
		if err := s.pg.CreateCourierProfile(ctx, &repository.CourierProfile{
			UserID: userID,
			Name:   in.CourierName,
			Phone:  in.CourierPhone,
		}); err != nil {
			return nil, fmt.Errorf("create courier profile: %w", err)
		}
	}

	return s.issueTokens(ctx, userID, in.Role)
}

type LoginResult struct {
	UserID       string
	Role         string
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.pg.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:       user.ID,
		Role:         user.Role,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) ValidateToken(_ context.Context, tokenStr string) (*Claims, error) {
	// ValidateToken не ходит в Redis - access token проверяется только по подписи.
	// access TTL = 15 мин, инвалидация через logout
	// применяется только к refresh токену.
	claims, err := s.jwt.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*RegisterResult, error) {
	// Шаг 1: проверить подпись и срок
	claims, err := s.jwt.ParseToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Шаг 2: убедиться что токен есть в Redis (не было logout)
	userID, err := s.rdb.GetUserIDByRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("check refresh token: %w", err)
	}

	// Дополнительная проверка: claims должны совпадать с userID из Redis
	if userID != claims.UserID {
		return nil, ErrInvalidCredentials
	}

	// Шаг 3: ротация - старый удаляем, выдаём новую пару
	if err := s.rdb.DeleteRefreshToken(ctx, userID, refreshTokenStr); err != nil {
		return nil, fmt.Errorf("delete old refresh token: %w", err)
	}

	return s.issueTokens(ctx, userID, claims.Role)
}

func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	claims, err := s.jwt.ParseToken(refreshTokenStr)
	if err != nil {
		// Токен иссяк - но логаут всё равно считаем успешным
		return nil
	}
	return s.rdb.DeleteRefreshToken(ctx, claims.UserID, refreshTokenStr)
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*repository.User, *repository.RestaurantProfile, *repository.CourierProfile, error) {
	user, err := s.pg.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, nil, ErrUserNotFound
		}
		return nil, nil, nil, fmt.Errorf("get user: %w", err)
	}

	switch user.Role {
	case "restaurant":
		rp, err := s.pg.GetRestaurantProfile(ctx, userID)
		if err != nil {
			return user, nil, nil, nil // профиль мог не создаться - не роняем сервис
		}
		return user, rp, nil, nil
	case "courier":
		cp, err := s.pg.GetCourierProfile(ctx, userID)
		if err != nil {
			return user, nil, nil, nil
		}
		return user, nil, cp, nil
	}

	return user, nil, nil, nil
}

// issueTokens - внутренний хелпер: генерирует пару токенов и сохраняет refresh в Redis
func (s *AuthService) issueTokens(ctx context.Context, userID, role string) (*RegisterResult, error) {
	accessToken, err := s.jwt.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.rdb.SaveRefreshToken(ctx, userID, refreshToken, s.jwt.RefreshTTL()); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &RegisterResult{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
