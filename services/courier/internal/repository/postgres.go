package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

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
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// Получение заказа по ID
func (r *PostgresRepo) GetOrderByID(ctx context.Context, orderID string) (*Order, error) {
	order := &Order{}
	query := `
		SELECT 
			o.id, 
			o.client_id, 
			o.restaurant_id, 
			o.courier_id, 
			o.address, 
			COALESCE(SUM(op.count * p.price), 0) AS total_price,
			o.status
		FROM orders o
		LEFT JOIN ordered_products op ON o.id = op.order_id
		LEFT JOIN products p ON op.product_id = p.id
		WHERE o.id = $1
		GROUP BY o.id`

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
		SELECT 
			o.id, 
			o.client_id, 
			o.restaurant_id, 
			o.courier_id, 
			o.address, 
			COALESCE(SUM(op.count * p.price), 0) AS total_price,
			o.status
		FROM orders o
		LEFT JOIN ordered_products op ON o.id = op.order_id
		LEFT JOIN products p ON op.product_id = p.id
		WHERE o.status = 'ready' AND o.courier_id IS NULL
		GROUP BY o.id
		ORDER BY o.id ASC
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
	query := `SELECT COUNT(*) FROM orders WHERE status = 'ready' AND courier_id IS NULL`

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
		SET courier_id = $1, status = 'delivering'
		WHERE id = $2 AND status = 'ready' AND courier_id IS NULL`

	result, err := r.db.ExecContext(ctx, query, courierID, orderID)
	if err != nil {
		return fmt.Errorf("assign courier: %w", err)
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

// Проверка, что заказ назначен этому курьеру
func (r *PostgresRepo) CheckOrderAssignedToCourier(ctx context.Context, orderID, courierID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1 AND courier_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, orderID, courierID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check order assigned: %w", err)
	}
	return exists, nil
}

// Все заказы курьера
func (r *PostgresRepo) ListOrdersByCourier(ctx context.Context, courierID string, limit, offset int32) ([]Order, error) {
	query := `
		SELECT 
			o.id, 
			o.client_id, 
			o.restaurant_id, 
			o.courier_id, 
			o.address, 
			COALESCE(SUM(op.count * p.price), 0) AS total_price,
			o.status
		FROM orders o
		LEFT JOIN ordered_products op ON o.id = op.order_id
		LEFT JOIN products p ON op.product_id = p.id
		WHERE o.courier_id = $1
		GROUP BY o.id
		ORDER BY o.id DESC
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
		  	AND status = 'delivering'
	`

	var count int32
	err := r.db.QueryRowContext(ctx, query, courierID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active orders: %w", err)
	}
	return count, nil
}

// Обновление статус заказа
func (r *PostgresRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

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
		SELECT op.id, op.order_id, op.product_id, p.name, op.count, p.price
		FROM ordered_products op
		JOIN products p ON op.product_id = p.id
		WHERE op.order_id = $1`

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
