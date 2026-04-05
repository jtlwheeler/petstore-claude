package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtlwheeler/petstore/internal/models"
)

// UserRepository provides access to user storage.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func scanUser(row pgx.Row) (models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.Username, &u.FirstName, &u.LastName,
		&u.Email, &u.Password, &u.Phone, &u.UserStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, first_name, last_name, email, password, phone, user_status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, username, first_name, last_name, email, password, phone, user_status`,
		user.Username, user.FirstName, user.LastName, user.Email, user.Password, user.Phone, user.UserStatus,
	).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Email, &user.Password, &user.Phone, &user.UserStatus,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// CreateBatch inserts multiple users and returns them.
func (r *UserRepository) CreateBatch(ctx context.Context, users []models.User) ([]models.User, error) {
	result := make([]models.User, 0, len(users))
	for _, u := range users {
		created, err := r.Create(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("creating user %s: %w", u.Username, err)
		}
		result = append(result, created)
	}
	return result, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (models.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, email, password, phone, user_status
		 FROM users WHERE username = $1`,
		username,
	)
	return scanUser(row)
}

// Update replaces a user's data by username.
func (r *UserRepository) Update(ctx context.Context, username string, user models.User) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET username = $1, first_name = $2, last_name = $3, email = $4,
		 password = $5, phone = $6, user_status = $7 WHERE username = $8`,
		user.Username, user.FirstName, user.LastName, user.Email,
		user.Password, user.Phone, user.UserStatus, username,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a user by username.
func (r *UserRepository) Delete(ctx context.Context, username string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Login returns the user if credentials match.
func (r *UserRepository) Login(ctx context.Context, username, password string) (models.User, error) {
	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		return models.User{}, err
	}
	if user.Password != password {
		return models.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
