package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

type Restaurant struct {
	pool *pgxpool.Pool
}

func NewRestaurant(pool *pgxpool.Pool) *Restaurant {
	return &Restaurant{pool}
}

func (r *Restaurant) AddProductIntoMenu(ctx context.Context, productInfo *models.ProductInfo) (uuid.UUID, error) {
	query := `
		INSERT INTO products(id, restaurant_id, name, description, price)
		VALUES($1, $2, $3, $4, $5);
    `
	productId := uuid.New()
	_, err := r.pool.Exec(
		ctx,
		query,
		productId,
		productInfo.RestaurantId,
		productInfo.Name,
		productInfo.Description,
		productInfo.Price,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return uuid.Nil, utils.ErrProductAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				return uuid.Nil, utils.ErrInvalidRestaurantID
			case pgerrcode.CheckViolation:
				return uuid.Nil, utils.ErrValidationFailed
			case pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
				return uuid.Nil, utils.ErrInvalidDataSchema
			case pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
				return uuid.Nil, utils.ErrConnectionFailure
			}
		}
		return uuid.Nil, fmt.Errorf("repository.DeleteProductFromMenu: %w", err)
	}

	return productId, nil
}

func (r *Restaurant) UpdateProductInMenu(ctx context.Context, product *models.FullProduct) (uuid.UUID, error) {
	query := `
	UPDATE products
	SET
    name = $2,
    description = $3,
    price = $4,
    WHERE id = $1 AND restaurant_id = $5;
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		product.Id,
		product.Name,
		product.Description,
		product.Price,
		product.RestaurantId,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return uuid.Nil, utils.ErrInvalidRestaurantID
			case pgerrcode.CheckViolation:
				return uuid.Nil, utils.ErrValidationFailed
			case pgerrcode.NotNullViolation:
				return uuid.Nil, utils.ErrValidationFailed
			case pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
				return uuid.Nil, utils.ErrInvalidDataSchema
			case pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
				return uuid.Nil, utils.ErrConnectionFailure
			}
		}
		return uuid.Nil, fmt.Errorf("repository.UpdateProductInMenu: %w", err)
	}

	return product.Id, nil
}

func (r *Restaurant) DeleteProductFromMenu(ctx context.Context, productId *models.ProductId) error {
	query := `
	DELETE FROM products
	WHERE id = $1;
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		productId.Id,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return utils.ErrProductHasDependencies
			case pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
				return utils.ErrInvalidDataSchema
			case pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
				return utils.ErrConnectionFailure
			}
		}
		return fmt.Errorf("repository.DeleteProductFromMenu: %w", err)
	}

	if result.RowsAffected() == 0 {
		return utils.ErrProductDoesNotFound
	}

	return nil
}

func (r *Restaurant) ListProducts(ctx context.Context, restaurantId *models.RestaurantId) ([]models.FullProduct, error) {
	query := `
	SELECT id, restaurant_id, name, description, price
	FROM products
	WHERE restaurant_id = $1;
	`

	rows, err := r.pool.Query(ctx, query, restaurantId.Id)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProducts: %w", err)
	}

	defer rows.Close()
	var products []models.FullProduct

	for rows.Next() {
		var product models.FullProduct
		if err := rows.Scan(
			&product.Id, &product.RestaurantId,
			&product.Name, &product.Description,
			&product.Price,
		); err != nil {
			return nil, fmt.Errorf("list products scan: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListProducts: %w", err)
	}

	if products == nil {
		products = []models.FullProduct{}
	}

	return products, nil
}

func (r *Restaurant) GetProduct(ctx context.Context, productId *models.ProductId) (*models.FullProduct, error) {
	query := `
	SELECT id, restaurant_id, name, description, price
    FROM products
	WHERE id = $1;
	`
	var product models.FullProduct
	err := r.pool.QueryRow(ctx, query, productId.Id).Scan(&product.Id, &product.RestaurantId, &product.Name, &product.Description, &product.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrProductDoesNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
				return nil, utils.ErrInvalidDataSchema
			case pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
				return nil, utils.ErrConnectionFailure
			}
		}

		return nil, fmt.Errorf("repository.GetProduct: %w", err)
	}

	return &product, nil
}

func (r *Restaurant) ChangeOrderStatus(ctx context.Context, order *models.OrderIdWithStatus) (uuid.UUID, error) {
	query := `
	UPDATE orders
    SET status = $1
	WHERE id = $2;
	`
	dbStatus, err := models.ConvertCommonOrderStatusToDBStatus(order.Status)
	if err != nil {
		return uuid.Nil, err
	}

	//TODO:через транзакцию, проверить, что статус у меняемого не cancelled и прошлый<предыдущи или cancelled
	_, err = r.pool.Exec(ctx, query, dbStatus, order.OrderId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return uuid.Nil, utils.ErrInvalidOrderID
			case pgerrcode.CheckViolation:
				return uuid.Nil, utils.ErrValidationFailed
			case pgerrcode.NotNullViolation:
				return uuid.Nil, utils.ErrValidationFailed
			case pgerrcode.UndefinedTable, pgerrcode.UndefinedColumn:
				return uuid.Nil, utils.ErrInvalidDataSchema
			case pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
				return uuid.Nil, utils.ErrConnectionFailure
			}
		}
		return uuid.Nil, fmt.Errorf("repository.ChangeOrderStatus: %w", err)
	}

	return order.OrderId, nil
}

func (r *Restaurant) ListOrders(ctx context.Context, restaurantId *models.RestaurantId) ([]models.Order, error) {
	query := `
	SELECT
    o.id AS order_id,
    o.client_id,
    o.courier_id,
    o.address,
    o.status,
    op.id AS ordered_product_id,
    op.count,
    p.id AS product_id,
    p.name AS product_name,
    p.price AS product_price,
    op.order_id AS order_id_of_product
	FROM ordered_products op
    	INNER JOIN orders o ON op.order_id = o.id AND o.restaurant_id = $1
    	INNER JOIN products p ON op.product_id = p.id
	ORDER BY op.id;
	`

	rows, err := r.pool.Query(ctx, query, restaurantId.Id)
	if err != nil {
		return nil, fmt.Errorf("repository.ListOrders query: %w", err)
	}

	defer rows.Close()
	ordersMap := make(map[uuid.UUID]*models.Order)

	for rows.Next() {
		var (
			orderID     uuid.UUID
			clientID    uuid.UUID
			courierID   *uuid.UUID
			address     string
			statusStr   string
			prodID      uuid.UUID
			prodQty     int32
			prodName    string
			prodPrice   int64
			prodOrderID uuid.UUID
		)

		if err := rows.Scan(
			&orderID, &clientID, &courierID, &address, &statusStr,
			&prodID, &prodQty, &prodName, &prodPrice, &prodOrderID,
		); err != nil {
			return nil, fmt.Errorf("repository.ListOrders scan: %w", err)
		}

		order, exists := ordersMap[orderID]
		if !exists {
			order = &models.Order{
				Id:           orderID,
				RestaurantId: restaurantId.Id,
				ClientId:     clientID,
				Address:      address,
				OrderStatus:  models.ConvertDBStatusToCommonOrderStatus(statusStr),
				OrderedItems: []models.OrderedProduct{},
			}
			if courierID != nil {
				order.CourierId = *courierID
			}
			ordersMap[orderID] = order
		}

		order.OrderedItems = append(order.OrderedItems, models.OrderedProduct{
			ProductId: prodID,
			Name:      prodName,
			OrderId:   prodOrderID,
			Price:     prodPrice,
			Quantity:  prodQty,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListRows iteration: %w", err)
	}

	orders := make([]models.Order, 0, len(ordersMap))
	for _, o := range ordersMap {
		var total int64
		for _, item := range o.OrderedItems {
			total += item.Price * int64(item.Quantity)
		}
		o.TotalPrice = total
		orders = append(orders, *o)
	}

	return orders, nil
}
