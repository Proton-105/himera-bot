package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Proton-105/himera-bot/internal/domain"
	usercache "github.com/Proton-105/himera-bot/internal/usercache"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	GetSettings(ctx context.Context, userID int64) (*domain.UserSettings, error)
	UpdateSettings(ctx context.Context, userID int64, settings *domain.UserSettings) error
	UpdateLastActiveAt(ctx context.Context, userID int64) error
	BlockUser(ctx context.Context, userID int64) error
	UnblockUser(ctx context.Context, userID int64) error
	IsBlocked(ctx context.Context, userID int64) (bool, error)
}

type userRepository struct {
	db    *sql.DB
	log   *slog.Logger
	cache *usercache.Cache
}

const userCacheTTL = 5 * time.Minute

// NewUserRepository creates a new SQL-backed user repository.
func NewUserRepository(db *sql.DB, log *slog.Logger, cache ...*usercache.Cache) UserRepository {
	var c *usercache.Cache
	if len(cache) > 0 {
		c = cache[0]
	}

	return &userRepository{
		db:    db,
		log:   log,
		cache: c,
	}
}

// FindByID retrieves a user from the database by their Telegram identifier.
func (r *userRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
		SELECT id, telegram_id, first_name, last_name, username, balance, last_active_at, is_blocked, created_at
		FROM users
		WHERE telegram_id = $1
	`

	if cached, err := r.getFromCache(ctx, id); err == nil && cached != nil {
		return cached, nil
	} else if err != nil {
		r.logCacheError("get", id, err)
	}

	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	if err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Balance,
		&user.LastActiveAt,
		&user.IsBlocked,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		r.logError("find_by_id", id, err)
		return nil, fmt.Errorf("select user by telegram id: %w", err)
	}

	if err := r.setCache(ctx, &user); err != nil {
		r.logCacheError("set", user.TelegramID, err)
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
		r.logError("create", user.TelegramID, err)
		return fmt.Errorf("insert user: %w", err)
	}

	if err := r.invalidateCache(ctx, user.TelegramID); err != nil {
		r.logCacheError("invalidate", user.TelegramID, err)
	}

	return nil
}

// GetSettings retrieves persisted user settings.
func (r *userRepository) GetSettings(ctx context.Context, userID int64) (*domain.UserSettings, error) {
	const query = `
		SELECT notifications_enabled, language, timezone
		FROM users_settings
		WHERE telegram_id = $1
	`

	var settings domain.UserSettings

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.NotificationsEnabled,
		&settings.Language,
		&settings.Timezone,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		r.logError("get_settings", userID, err)
		return nil, fmt.Errorf("select user settings: %w", err)
	}

	return &settings, nil
}

// UpdateSettings creates or updates user settings atomically.
func (r *userRepository) UpdateSettings(ctx context.Context, userID int64, settings *domain.UserSettings) error {
	const query = `
		INSERT INTO users_settings (telegram_id, notifications_enabled, language, timezone)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id) DO UPDATE
		SET notifications_enabled = EXCLUDED.notifications_enabled,
			language = EXCLUDED.language,
			timezone = EXCLUDED.timezone,
			updated_at = NOW()
	`

	if _, err := r.db.ExecContext(ctx, query, userID, settings.NotificationsEnabled, settings.Language, settings.Timezone); err != nil {
		r.logError("update_settings", userID, err)
		return fmt.Errorf("upsert user settings: %w", err)
	}

	return nil
}

// UpdateLastActiveAt refreshes the last activity timestamp for a user.
func (r *userRepository) UpdateLastActiveAt(ctx context.Context, userID int64) error {
	const query = `
		UPDATE users
		SET last_active_at = NOW()
		WHERE telegram_id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		r.logError("update_last_active_at", userID, err)
		return fmt.Errorf("update last_active_at: %w", err)
	}

	if err := r.invalidateCache(ctx, userID); err != nil {
		r.logCacheError("invalidate", userID, err)
	}

	return nil
}

// BlockUser marks a user as blocked.
func (r *userRepository) BlockUser(ctx context.Context, userID int64) error {
	const query = `
		UPDATE users
		SET is_blocked = TRUE
		WHERE telegram_id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		r.logError("block_user", userID, err)
		return fmt.Errorf("block user: %w", err)
	}

	if err := r.invalidateCache(ctx, userID); err != nil {
		r.logCacheError("invalidate", userID, err)
	}

	return nil
}

// UnblockUser removes block flag from a user.
func (r *userRepository) UnblockUser(ctx context.Context, userID int64) error {
	const query = `
		UPDATE users
		SET is_blocked = FALSE
		WHERE telegram_id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		r.logError("unblock_user", userID, err)
		return fmt.Errorf("unblock user: %w", err)
	}

	if err := r.invalidateCache(ctx, userID); err != nil {
		r.logCacheError("invalidate", userID, err)
	}

	return nil
}

// IsBlocked indicates whether a user is blocked.
func (r *userRepository) IsBlocked(ctx context.Context, userID int64) (bool, error) {
	const query = `
		SELECT is_blocked
		FROM users
		WHERE telegram_id = $1
	`

	var blocked bool
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&blocked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, sql.ErrNoRows
		}

		r.logError("is_blocked", userID, err)
		return false, fmt.Errorf("select is_blocked: %w", err)
	}

	return blocked, nil
}

func (r *userRepository) logError(operation string, userID int64, err error) {
	if r.log == nil {
		return
	}

	r.log.Error(
		"user repository operation failed",
		slog.String("operation", operation),
		slog.Int64("telegram_id", userID),
		slog.Any("error", err),
	)
}

func (r *userRepository) logCacheError(operation string, userID int64, err error) {
	if err == nil || r.log == nil {
		return
	}

	r.log.Warn(
		"user cache operation failed",
		slog.String("operation", operation),
		slog.Int64("telegram_id", userID),
		slog.Any("error", err),
	)
}

func (r *userRepository) getFromCache(ctx context.Context, userID int64) (*domain.User, error) {
	if r.cache == nil {
		return nil, nil
	}

	return r.cache.Get(ctx, userID)
}

func (r *userRepository) setCache(ctx context.Context, user *domain.User) error {
	if r.cache == nil || user == nil {
		return nil
	}

	return r.cache.Set(ctx, user.TelegramID, user, userCacheTTL)
}

func (r *userRepository) invalidateCache(ctx context.Context, userID int64) error {
	if r.cache == nil {
		return nil
	}

	return r.cache.Invalidate(ctx, userID)
}
