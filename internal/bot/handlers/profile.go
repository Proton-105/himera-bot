package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/user"
)

// NewProfileHandler returns a handler for the /profile command.
func NewProfileHandler(userService *user.Service, log *slog.Logger) Handler {
	return func(c telebot.Context) error {
		if c == nil {
			return nil
		}

		sender := c.Sender()
		if sender == nil {
			return c.Send("Unable to load your profile right now.")
		}

		ctx := context.Background()
		profile, err := userService.GetOrCreate(ctx, sender)
		if err != nil {
			if log != nil {
				log.Error("profile handler failed to fetch user", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return c.Send("Unable to load your profile right now. Please try again later.")
		}

		username := strings.TrimSpace(profile.Username)
		switch {
		case username == "":
			username = fmt.Sprintf("ID:%d", sender.ID)
		case !strings.HasPrefix(username, "@"):
			username = "@" + username
		}

		balanceUSD := float64(profile.Balance) / 100
		message := fmt.Sprintf(
			"Username: %s\nBalance: %.2f USD\nJoined: %s",
			username,
			balanceUSD,
			profile.CreatedAt.Format("January 2, 2006"),
		)

		return c.Send(message)
	}
}
