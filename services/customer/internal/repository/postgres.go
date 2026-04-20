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

type Order struct {
	ID           string
	UserID       string
	RestaurantID string
	CourierID    sql.NullString
	Address      string
	TotalPrice   int64
	Status       string
}

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Name      string
	Quantity  int32
	Price     int64
}

type OrderWithItems struct {
	Order Order
	Items []OrderItem
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

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// Создание заказа
func (r *PostgresRepo) CreateOrder(ctx context.Context, order *Order, items []OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Вставляем заказ
	orderQuery := `
		INSERT INTO orders (id, user_id, restaurant_id, address, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID,
		order.UserID,
		order.RestaurantID,
		order.Address,
		order.TotalPrice,
		order.Status,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, name, quantity, price)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for _, item := range items {
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID,
			order.ID,
			item.ProductID,
			item.Name,
			item.Quantity,
			item.Price,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit()
}

// Получение заказа по ID
func (r *PostgresRepo) GetOrderByID(ctx context.Context, orderID string) (*Order, error) {
	order := &Order{}
	query := `
		SELECT id, user_id, restaurant_id, courier_id, address, total_price, status, created_at, updated_at
		FROM orders
		WHERE id = $1`

	var courierID sql.NullString
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.RestaurantID,
		&courierID,
		&order.Address,
		&order.TotalPrice,
		&order.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	order.CourierID = courierID
	return order, nil
}

// Получение позиций заказа
func (r *PostgresRepo) GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, name, quantity, price
		FROM order_items
		WHERE order_id = $1`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Name, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// Получение заказа вместе с позициями
func (r *PostgresRepo) GetOrderWithItems(ctx context.Context, orderID string) (*OrderWithItems, error) {
	order, err := r.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items, err := r.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &OrderWithItems{
		Order: *order,
		Items: items,
	}, nil
}

// Обновление статуса заказа
func (r *PostgresRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Назначение курьера на заказ
func (r *PostgresRepo) UpdateOrderCourier(ctx context.Context, orderID, courierID string) error {
	query := `
		UPDATE orders
		SET courier_id = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, courierID, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("update order courier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Отмена заказа
func (r *PostgresRepo) CancelOrder(ctx context.Context, orderID string) error {
	return r.UpdateOrderStatus(ctx, orderID, "cancelled")
}

// Возвращает заказы пользователя с пагинацией
func (r *PostgresRepo) ListOrdersByUserID(ctx context.Context, userID string, limit, offset int32) ([]Order, error) {
	query := `
		SELECT o.id, COALESCE(r.name, ''), o.status, o.total_price, o.created_at
		FROM orders o
		LEFT JOIN restaurants r ON o.restaurant_id = r.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.Status, &o.TotalPrice); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// Общее количество заказов пользователя
func (r *PostgresRepo) CountOrdersByUserID(ctx context.Context, userID string) (int32, error) {
	query := `SELECT COUNT(*) FROM orders WHERE user_id = $1`

	var count int32
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count orders: %w", err)
	}

	return count, nil
}

// Проверка, что заказ принадлежит пользователю
func (r *PostgresRepo) CheckOrderBelongsToUser(ctx context.Context, orderID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check order belongs to user: %w", err)
	}

	return exists, nil
}
