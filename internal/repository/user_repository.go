package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Proton-105/himera-bot/internal/domain"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type userRepository struct {
	db  *sql.DB
	log *slog.Logger
}

// NewUserRepository creates a new SQL-backed user repository.
func NewUserRepository(db *sql.DB, log *slog.Logger) UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}

// FindByID retrieves a user from the database by their Telegram identifier.
func (r *userRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, username, balance, created_at
		FROM users
		WHERE telegram_id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	if err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Balance,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		if r.log != nil {
			r.log.Error("failed to fetch user by telegram id", slog.Int64("telegram_id", id), slog.Any("error", err))
		}
		return nil, fmt.Errorf("select user by telegram id: %w", err)
	}

	return &user, nil
}

// Create persists a new user record in the database.
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (telegram_id, first_name, last_name, username, balance, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Balance,
		user.CreatedAt,
	); err != nil {
		if r.log != nil {
			r.log.Error("failed to create user", slog.Int64("telegram_id", user.TelegramID), slog.Any("error", err))
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}
