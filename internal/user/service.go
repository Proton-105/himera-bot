package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/domain"
	"github.com/Proton-105/himera-bot/internal/repository"
)

// Service provides business operations over users.
type Service struct {
	repo repository.UserRepository
	log  *slog.Logger
}

// NewService constructs a new Service instance.
func NewService(repo repository.UserRepository, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// GetOrCreate fetches a user by telegram ID or creates a new profile when missing.
func (s *Service) GetOrCreate(ctx context.Context, telegramUser *telebot.User) (*domain.User, error) {
	if telegramUser == nil {
		return nil, errors.New("telegram user is nil")
	}

	user, err := s.repo.FindByID(ctx, telegramUser.ID)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		s.logError("get_or_create.find", telegramUser.ID, err)
		return nil, fmt.Errorf("get user: %w", err)
	}

	now := time.Now().UTC()
	newUser := &domain.User{
		TelegramID:   telegramUser.ID,
		FirstName:    telegramUser.FirstName,
		LastName:     telegramUser.LastName,
		Username:     telegramUser.Username,
		Balance:      0,
		LastActiveAt: now,
		CreatedAt:    now,
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logError("get_or_create.create", telegramUser.ID, err)
		return nil, fmt.Errorf("create user: %w", err)
	}

	return newUser, nil
}

// GetSettings returns persisted settings for the supplied user.
func (s *Service) GetSettings(ctx context.Context, userID int64) (*domain.UserSettings, error) {
	settings, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		s.logError("get_settings", userID, err)
		return nil, err
	}

	return settings, nil
}

// UpdateSettings saves user preferences.
func (s *Service) UpdateSettings(ctx context.Context, userID int64, settings *domain.UserSettings) error {
	if err := s.repo.UpdateSettings(ctx, userID, settings); err != nil {
		s.logError("update_settings", userID, err)
		return err
	}

	return nil
}

// UpdateLastActive refreshes the last_active_at field for the user.
func (s *Service) UpdateLastActive(ctx context.Context, userID int64) error {
	if err := s.repo.UpdateLastActiveAt(ctx, userID); err != nil {
		s.logError("update_last_active", userID, err)
		return err
	}

	return nil
}

func (s *Service) logError(operation string, telegramID int64, err error) {
	if s == nil || s.log == nil || err == nil {
		return
	}

	s.log.Error("user service operation failed",
		slog.String("operation", operation),
		slog.Int64("telegram_id", telegramID),
		slog.Any("error", err),
	)
}
