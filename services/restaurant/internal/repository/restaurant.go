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
		if err := rows.Scan(&product.Id, &product.RestaurantId, &product.Name, &product.Description, &product.Price); err != nil {
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
