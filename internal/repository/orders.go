package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/riouske/gophermart/internal/model"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderExists        = errors.New("order already exists for this user")
	ErrOrderExistsForUser = errors.New("order already exists for another user")
)

type OrderRepositoryInterface interface {
	Create(order *model.Order) error
	GetByID(id int64) (*model.Order, error)
	GetByNumber(number string) (*model.Order, error)
	GetByUserID(userID int64) ([]*model.Order, error)
	UpdateStatus(id int64, status model.OrderStatus) error
	UpdateAccrual(id int64, accrual float64, status model.OrderStatus) error
}

type OrderRepository struct {
	Impl OrderRepositoryInterface
	db   *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	repo := &OrderRepository{db: db}
	repo.Impl = &PostgresOrderRepository{db: db}
	return repo
}

// Create delegates to the implementation
func (r *OrderRepository) Create(order *model.Order) error {
	return r.Impl.Create(order)
}

// GetByID delegates to the implementation
func (r *OrderRepository) GetByID(id int64) (*model.Order, error) {
	return r.Impl.GetByID(id)
}

// GetByNumber delegates to the implementation
func (r *OrderRepository) GetByNumber(number string) (*model.Order, error) {
	return r.Impl.GetByNumber(number)
}

// GetByUserID delegates to the implementation
func (r *OrderRepository) GetByUserID(userID int64) ([]*model.Order, error) {
	return r.Impl.GetByUserID(userID)
}

// UpdateStatus delegates to the implementation
func (r *OrderRepository) UpdateStatus(id int64, status model.OrderStatus) error {
	return r.Impl.UpdateStatus(id, status)
}

// UpdateAccrual delegates to the implementation
func (r *OrderRepository) UpdateAccrual(id int64, accrual float64, status model.OrderStatus) error {
	return r.Impl.UpdateAccrual(id, accrual, status)
}

// PostgresOrderRepository is the PostgreSQL implementation of OrderRepositoryInterface
type PostgresOrderRepository struct {
	db *sql.DB
}

// Create adds a new order to the database
func (r *PostgresOrderRepository) Create(order *model.Order) error {
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id`

	order.UploadedAt = time.Now()

	err := r.db.QueryRow(
		query,
		order.UserID,
		order.Number,
		order.Status,
		order.UploadedAt,
	).Scan(&order.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Check if the order exists for the same user
			existingOrder, err := r.GetByNumber(order.Number)
			if err == nil && existingOrder.UserID == order.UserID {
				return ErrOrderExists
			}
			return ErrOrderExistsForUser
		}
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by its ID
func (r *PostgresOrderRepository) GetByID(id int64) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
              FROM orders 
              WHERE id = $1`

	order := &model.Order{}
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetByNumber retrieves an order by its number
func (r *PostgresOrderRepository) GetByNumber(number string) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
              FROM orders 
              WHERE number = $1`

	order := &model.Order{}
	err := r.db.QueryRow(query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetByUserID retrieves all orders for a specific user
func (r *PostgresOrderRepository) GetByUserID(userID int64) ([]*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
              FROM orders 
              WHERE user_id = $1
              ORDER BY uploaded_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders rows: %w", err)
	}

	return orders, nil
}

// UpdateStatus updates the status of an order
func (r *PostgresOrderRepository) UpdateStatus(id int64, status model.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// UpdateAccrual updates the accrual amount for an order
func (r *PostgresOrderRepository) UpdateAccrual(id int64, accrual float64, status model.OrderStatus) error {
	query := `UPDATE orders SET accrual = $1, status = $2 WHERE id = $3`

	result, err := r.db.Exec(query, accrual, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order accrual: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}