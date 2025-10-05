package repositories

import (
	"be/internal/models"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrOutOfStock = errors.New("OUT_OF_STOCK")

type OrderRepository interface {
	CreateOrderWithStock(ctx context.Context, productID int64, quantity int, buyerID string) (*models.Order, error)
	GetByID(ctx context.Context, id int64) (*models.Order, error)
	ResetProductStock(ctx context.Context, productID int64, stock int) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository { return &orderRepository{db: db} }

// CreateOrderWithStock decrements stock atomically and creates an order.
func (r *orderRepository) CreateOrderWithStock(ctx context.Context, productID int64, quantity int, buyerID string) (*models.Order, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Atomically decrement stock if enough is available
	res, err := tx.ExecContext(ctx, `
		UPDATE products
		SET stock = stock - $1,
			updated_at = now()
		WHERE id = $2
		  AND stock >= $1
	`, quantity, productID)
	if err != nil {
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return nil, ErrOutOfStock
	}

	// Retrieve price for order creation
	var priceCents int64
	if err = tx.QueryRowContext(ctx, `SELECT price_cents FROM products WHERE id = $1`, productID).Scan(&priceCents); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	total := priceCents * int64(quantity)

	var orderID int64
	if err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (product_id, buyer_id, quantity, total_cents, status)
		VALUES ($1, $2, $3, $4, 'CREATED')
		RETURNING id
	`, productID, buyerID, quantity, total).Scan(&orderID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Order{
		ID:         orderID,
		ProductID:  productID,
		BuyerID:    buyerID,
		Quantity:   quantity,
		TotalCents: total,
		Status:     "CREATED",
		CreatedAt:  time.Now(),
	}, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	row := r.db.QueryRowxContext(ctx, `SELECT id, product_id, buyer_id, quantity, total_cents, status, created_at FROM orders WHERE id = $1`, id)
	var o models.Order
	if err := row.Scan(&o.ID, &o.ProductID, &o.BuyerID, &o.Quantity, &o.TotalCents, &o.Status, &o.CreatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) ResetProductStock(ctx context.Context, productID int64, stock int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET stock=$1, updated_at=now() WHERE id=$2`, stock, productID)
	return err
}
