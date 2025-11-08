package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/internal/idempotency"
)

// Idempotency ensures handlers execute at most once per Telegram update key.
func Idempotency(manager idempotency.Manager, log *slog.Logger) handlers.Middleware {
	if manager == nil {
		return func(next handlers.Handler) handlers.Handler {
			return next
		}
	}
	if log == nil {
		log = slog.Default()
	}

	return func(next handlers.Handler) handlers.Handler {
		if next == nil {
			return nil
		}

		return func(c telebot.Context) error {
			key := extractIdempotencyKey(c)
			if key == "" {
				return next(c)
			}

			ctx := context.Background()

			result, err := manager.Execute(ctx, key, 24*time.Hour, func(execCtx context.Context) (interface{}, error) {
				return nil, next(c)
			})
			if err != nil {
				if errors.Is(err, idempotency.ErrRequestInProgress) {
					return nil
				}

				log.Error("idempotent handler failed", slog.String("key", key), slog.Any("error", err))
				return err
			}

			if result != nil && result.FromCache {
				return nil
			}

			return nil
		}
	}
}

func extractIdempotencyKey(c telebot.Context) string {
	if c == nil {
		return ""
	}

	if cb := c.Callback(); cb != nil {
		if cb.ID != "" {
			return fmt.Sprintf("cb:%s", cb.ID)
		}

		if cb.Message != nil {
			chatID := int64(0)
			if cb.Message.Chat != nil {
				chatID = cb.Message.Chat.ID
			}
			return fmt.Sprintf("cb-msg:%d:%d", chatID, cb.Message.ID)
		}
	}

	if msg := c.Message(); msg != nil {
		chatID := int64(0)
		if msg.Chat != nil {
			chatID = msg.Chat.ID
		}
		if msg.ID != 0 {
			return fmt.Sprintf("msg:%d:%d", chatID, msg.ID)
		}
	}

	return ""
}
