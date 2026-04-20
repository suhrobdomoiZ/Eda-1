package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
}

type RestaurantProfile struct {
	UserID  string
	Name    string
	Address string
	Phone   string
}

type CourierProfile struct {
	UserID string
	Name   string
	Phone  string
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) CreateUser(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (id, username, password_hash, role)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.PasswordHash, u.Role)
	if err != nil {
		if isPgUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	query := `SELECT id, username, password_hash, role FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (r *PostgresRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	query := `SELECT id, username, password_hash, role FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *PostgresRepo) CreateRestaurantProfile(ctx context.Context, p *RestaurantProfile) error {
	query := `
		INSERT INTO restaurant_profiles (user_id, name, address, phone)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, p.UserID, p.Name, p.Address, p.Phone)
	if err != nil {
		return fmt.Errorf("create restaurant profile: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetRestaurantProfile(ctx context.Context, userID string) (*RestaurantProfile, error) {
	p := &RestaurantProfile{}
	query := `SELECT user_id, name, address, phone FROM restaurant_profiles WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&p.UserID, &p.Name, &p.Address, &p.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get restaurant profile: %w", err)
	}
	return p, nil
}

func (r *PostgresRepo) CreateCourierProfile(ctx context.Context, p *CourierProfile) error {
	query := `
		INSERT INTO courier_profiles (user_id, name, phone)
		VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, p.UserID, p.Name, p.Phone)
	if err != nil {
		return fmt.Errorf("create courier profile: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetCourierProfile(ctx context.Context, userID string) (*CourierProfile, error) {
	p := &CourierProfile{}
	query := `SELECT user_id, name, phone FROM courier_profiles WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&p.UserID, &p.Name, &p.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get courier profile: %w", err)
	}
	return p, nil
}

// isPgUniqueViolation проверяет pgcode без импорта pq напрямую
func isPgUniqueViolation(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		contains(err.Error(), "23505")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
