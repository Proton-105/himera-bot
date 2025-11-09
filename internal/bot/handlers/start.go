package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Proton-105/himera-bot/internal/state"
	"gopkg.in/telebot.v3"
)

// RegisterStartHandler registers the /start command handler that manages user state via the FSM.
func RegisterStartHandler(bot *telebot.Bot, fsm state.StateMachine, log *slog.Logger) {
	if bot == nil || fsm == nil {
		return
	}

	handler := func(c telebot.Context) error {
		sender := c.Sender()
		if sender == nil {
			return c.Send("An internal error occurred. Please try again later.")
		}

		ctx := context.Background()
		userID := sender.ID

		_, err := fsm.GetState(ctx, userID)
		switch {
		case err == nil:
			return c.Send("Welcome back!")
		case errors.Is(err, state.ErrStateNotFound):
			if setErr := fsm.SetState(ctx, userID, state.StateIdle, nil); setErr != nil {
				if log != nil {
					log.Error("failed to set initial user state", slog.Int64("telegram_id", userID), slog.Any("error", setErr))
				}
				return c.Send("An internal error occurred. Please try again later.")
			}
			return c.Send("Welcome! Let's get you set up.")
		default:
			if log != nil {
				log.Error("failed to fetch user state", slog.Int64("telegram_id", userID), slog.Any("error", err))
			}
			return c.Send("An internal error occurred. Please try again later.")
		}
	}

	bot.Handle("/start", handler)
}
