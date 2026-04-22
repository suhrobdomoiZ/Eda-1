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

// Заказы со статусом 'ready' без назначенного курьера
func (r *PostgresRepo) ListAvailableOrders(ctx context.Context, limit, offset int32) ([]Order, error) {
	query := `
		SELECT id, user_id, restaurant_id, courier_id, address, total_price, status, created_at, updated_at
		FROM orders
		WHERE status = 'ready' AND (courier_id IS NULL OR courier_id = '')
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query available orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		var courierID sql.NullString
		if err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.RestaurantID,
			&courierID,
			&o.Address,
			&o.TotalPrice,
			&o.Status,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CourierID = courierID
		orders = append(orders, o)
	}

	return orders, nil
}

// Количество доступных заказов
func (r *PostgresRepo) CountAvailableOrders(ctx context.Context) (int32, error) {
	query := `SELECT COUNT(*) FROM orders WHERE status = 'ready' AND (courier_id IS NULL OR courier_id = '')`

	var count int32
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count available orders: %w", err)
	}

	return count, nil
}

// Назначает курьера на заказ и меняет статус на 'delivering'
func (r *PostgresRepo) AssignCourierToOrder(ctx context.Context, orderID, courierID string) error {
	query := `
		UPDATE orders
		SET courier_id = $1, status = 'delivering', updated_at = $2
		WHERE id = $3 AND status = 'ready' AND (courier_id IS NULL OR courier_id = '')`

	result, err := r.db.ExecContext(ctx, query, courierID, orderID)
	if err != nil {
		return fmt.Errorf("assign courier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound // Заказ уже занят или не в статусе ready
	}

	return nil
}

// Проверка, что заказ назначен этому курьеру
func (r *PostgresRepo) CheckOrderAssignedToCourier(ctx context.Context, orderID, courierID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1 AND courier_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, orderID, courierID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check order assigned to courier: %w", err)
	}

	return exists, nil
}

// Все заказы курьера
func (r *PostgresRepo) ListOrdersByCourier(ctx context.Context, courierID string, limit, offset int32) ([]Order, error) {
	query := `
		SELECT id, user_id, restaurant_id, courier_id, address, total_price, status, created_at, updated_at
		FROM orders
		WHERE courier_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, courierID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query courier orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		var cID sql.NullString
		if err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.RestaurantID,
			&cID,
			&o.Address,
			&o.TotalPrice,
			&o.Status,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CourierID = cID
		orders = append(orders, o)
	}

	return orders, nil
}

// Общее количество заказов курьера
func (r *PostgresRepo) CountOrdersByCourier(ctx context.Context, courierID string) (int32, error) {
	query := `SELECT COUNT(*) FROM orders WHERE courier_id = $1`

	var count int32
	err := r.db.QueryRowContext(ctx, query, courierID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count courier orders: %w", err)
	}

	return count, nil
}

// Количество активных заказов курьера
func (r *PostgresRepo) CountActiveOrdersByCourier(ctx context.Context, courierID string) (int32, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
		WHERE courier_id = $1
		  AND status IN ('assigned', 'picked_up', 'delivering')`

	var count int32
	err := r.db.QueryRowContext(ctx, query, courierID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active orders: %w", err)
	}

	return count, nil
}

// Обновление статус заказа
func (r *PostgresRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, orderID)
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

// Получение позиций заказа
type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Name      string
	Quantity  int32
	Price     int64
}

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
