package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/internal/domain"
	errors "github.com/Proton-105/himera-bot/internal/errors"
	"github.com/Proton-105/himera-bot/internal/repository"
	"github.com/Proton-105/himera-bot/internal/user"
)

const defaultInitialBalance int64 = 10000

// RecoveryMiddleware catches panics, reports them via the centralized handler, and notifies the user.
func RecoveryMiddleware(log *slog.Logger, errHandler *errors.Handler) handlers.Middleware {
	if log == nil {
		log = slog.Default()
	}

	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic recovered in handler", slog.Any("panic", r), slog.String("stack", string(debug.Stack())))

					userMsg := "⚠️ Something went wrong. Please try again later."
					if errHandler != nil {
						appErr := errors.NewDatabaseError(fmt.Errorf("panic recovered: %v", r))
						if msg, _ := errHandler.Handle(context.Background(), appErr); msg != "" {
							userMsg = msg
						}
					}

					if c != nil {
						if sendErr := c.Send(userMsg); sendErr != nil {
							log.Error("failed to notify user about panic", slog.Any("error", sendErr))
						}
					}

					err = nil
				}
			}()

			return next(c)
		}
	}
}

// ErrorHandlingMiddleware centralizes error reporting and user messaging for handler failures.
func ErrorHandlingMiddleware(errHandler *errors.Handler) handlers.Middleware {
	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			userMsg := "Произошла ошибка. Попробуйте позже"
			if errHandler != nil {
				if msg, _ := errHandler.Handle(context.Background(), err); msg != "" {
					userMsg = msg
				}
			}

			if c != nil {
				_ = c.Send(userMsg)
			}

			return nil
		}
	}
}

// LoggingMiddleware logs basic telemetry about incoming updates.
func LoggingMiddleware(log *slog.Logger) handlers.Middleware {
	if log == nil {
		log = slog.Default()
	}

	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) error {
			start := time.Now()
			userID := int64(0)
			if c != nil && c.Sender() != nil {
				userID = c.Sender().ID
			}

			action := ""
			if c != nil {
				if cb := c.Callback(); cb != nil {
					action = cb.Data
				} else {
					action = c.Text()
				}
			}

			log.Info("handling update", slog.Int64("user_id", userID), slog.String("action", action))
			err := next(c)
			log.Info("handled update",
				slog.Int64("user_id", userID),
				slog.String("action", action),
				slog.Duration("duration", time.Since(start)),
				slog.Any("error", err),
			)

			return err
		}
	}
}

// AuthMiddleware ensures that each incoming request is associated with a user record.
func AuthMiddleware(userRepo repository.UserRepository, log *slog.Logger) handlers.Middleware {
	if log == nil {
		log = slog.Default()
	}

	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) error {
			if userRepo == nil || c == nil || c.Sender() == nil {
				return next(c)
			}

			ctx := context.Background()
			userID := c.Sender().ID

			user, err := userRepo.FindByID(ctx, userID)
			if err != nil {
				if err == sql.ErrNoRows {
					newUser := &domain.User{
						TelegramID: userID,
						FirstName:  c.Sender().FirstName,
						LastName:   c.Sender().LastName,
						Username:   c.Sender().Username,
						Balance:    defaultInitialBalance,
						CreatedAt:  time.Now().UTC(),
					}

					if createErr := userRepo.Create(ctx, newUser); createErr != nil {
						log.Error("failed to create user", slog.Int64("user_id", userID), slog.Any("error", createErr))
						return createErr
					}

					log.Info("created new user", slog.Int64("user_id", userID))
				} else {
					log.Error("failed to find user", slog.Int64("user_id", userID), slog.Any("error", err))
					return err
				}
			} else if user == nil {
				log.Warn("user repository returned nil user without error", slog.Int64("user_id", userID))
			}

			return next(c)
		}
	}
}

// LastActiveMiddleware records user activity timestamps without blocking request flow.
func LastActiveMiddleware(userService *user.Service) handlers.Middleware {
	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) error {
			if userService != nil && c != nil && c.Sender() != nil {
				userID := c.Sender().ID

				go func(id int64) {
					ctx := context.Background()
					_ = userService.UpdateLastActive(ctx, id)
				}(userID)
			}

			return next(c)
		}
	}
}
