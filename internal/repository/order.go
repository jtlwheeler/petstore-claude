package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtlwheeler/petstore/internal/models"
)

// OrderRepository provides access to order storage.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create inserts a new order and returns it with its assigned ID.
func (r *OrderRepository) Create(ctx context.Context, order models.Order) (models.Order, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO orders (pet_id, quantity, ship_date, status, complete)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, pet_id, quantity, ship_date, status, complete`,
		order.PetID, order.Quantity, order.ShipDate, order.Status, order.Complete,
	).Scan(
		&order.ID, &order.PetID, &order.Quantity, &order.ShipDate, &order.Status, &order.Complete,
	)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

// GetByID retrieves an order by its ID.
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (models.Order, error) {
	var order models.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, pet_id, quantity, ship_date, status, complete FROM orders WHERE id = $1`,
		id,
	).Scan(
		&order.ID, &order.PetID, &order.Quantity, &order.ShipDate, &order.Status, &order.Complete,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}
		return models.Order{}, err
	}
	return order, nil
}

// Delete removes an order by ID.
func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetInventory returns a map of pet status to count.
func (r *OrderRepository) GetInventory(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM pets GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inventory := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		inventory[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return inventory, nil
}
