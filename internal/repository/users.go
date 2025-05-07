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
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (login, password, created_at, updated_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRow(query, user.Login, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByLogin(login string) (*model.User, error) {
	query := `SELECT id, login, password, created_at, updated_at, deleted_at 
              FROM users 
              WHERE login = $1 AND deleted_at IS NULL`

	user := &model.User{}
	err := r.db.QueryRow(query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	query := `SELECT id, login, password, created_at, updated_at, deleted_at 
              FROM users 
              WHERE id = $1 AND deleted_at IS NULL`

	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
