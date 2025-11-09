package handlers

import (
	"context"
	"log/slog"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/state"
)

// NewCancelHandler resets user state and returns the user to the main menu.
func NewCancelHandler(fsm state.StateMachine, kb *keyboard.Builder, log *slog.Logger) Handler {
	if log == nil {
		log = slog.Default()
	}

	return func(c telebot.Context) error {
		if c == nil || c.Sender() == nil {
			log.Warn("cancel handler invoked without sender context")
			return nil
		}

		if fsm == nil {
			log.Error("state machine is not configured for cancel handler")
			return nil
		}

		ctx := context.Background()
		userID := c.Sender().ID

		if err := fsm.ClearState(ctx, userID); err != nil {
			log.Error("failed to clear user state", slog.Int64("user_id", userID), slog.Any("error", err))
			return err
		}

		if err := c.Send("Operation cancelled. Returning to main menu."); err != nil {
			log.Error("failed to notify user about cancellation", slog.Int64("user_id", userID), slog.Any("error", err))
			return err
		}

		if kb == nil {
			log.Warn("keyboard builder is not configured for cancel handler")
			return nil
		}

		if err := c.Send("Main menu:", kb.MainMenu()); err != nil {
			log.Error("failed to send main menu", slog.Int64("user_id", userID), slog.Any("error", err))
			return err
		}

		return nil
	}
}
