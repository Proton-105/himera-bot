package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/domain"
	"github.com/Proton-105/himera-bot/internal/user"
)

const (
	settingsToggleNotificationsData = "settings_toggle_notifications"
	settingsLanguageDataPrefix      = "settings_set_language_"
)

// NewSettingsHandler returns the /settings command handler.
func NewSettingsHandler(userService *user.Service, kb *keyboard.Builder, log *slog.Logger) Handler {
	return func(c telebot.Context) error {
		if c == nil {
			return nil
		}

		if userService == nil {
			return c.Send("Settings are temporarily unavailable.")
		}

		sender := c.Sender()
		if sender == nil {
			return c.Send("Unable to load your settings right now.")
		}

		ctx := context.Background()
		settings, err := userService.GetSettings(ctx, sender.ID)
		switch {
		case err == nil:
			// ok
		case errors.Is(err, sql.ErrNoRows):
			settings = defaultUserSettings()
		default:
			if log != nil {
				log.Error("settings handler: failed to fetch settings", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return c.Send("Unable to load your settings right now. Please try again later.")
		}

		message := fmt.Sprintf(
			"Notifications: %s\nLanguage: %s\nTimezone: %s",
			boolLabel(settings.NotificationsEnabled, "On", "Off"),
			strings.ToUpper(settings.Language),
			settings.Timezone,
		)

		markup := buildSettingsKeyboard(kb, settings)

		return c.Send(message, markup)
	}
}

// HandleToggleNotifications returns a callback handler that toggles notification preference.
func HandleToggleNotifications(userService *user.Service, log *slog.Logger) CallbackHandler {
	return func(c telebot.Context) error {
		if c == nil || userService == nil {
			return nil
		}

		sender := c.Sender()
		if sender == nil {
			return respondCallback(c, "User not found", true)
		}

		ctx := context.Background()
		settings, err := userService.GetSettings(ctx, sender.ID)
		switch {
		case err == nil:
		case errors.Is(err, sql.ErrNoRows):
			settings = defaultUserSettings()
		default:
			if log != nil {
				log.Error("toggle notifications: failed to load settings", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return respondCallback(c, "Unable to update settings", true)
		}

		settings.NotificationsEnabled = !settings.NotificationsEnabled
		if err := userService.UpdateSettings(ctx, sender.ID, settings); err != nil {
			if log != nil {
				log.Error("toggle notifications: failed to save settings", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return respondCallback(c, "Unable to update settings", true)
		}

		statusText := boolLabel(settings.NotificationsEnabled, "Notifications enabled", "Notifications disabled")
		return respondCallback(c, statusText, false)
	}
}

// HandleSetLanguage returns a callback handler that updates the language preference.
func HandleSetLanguage(userService *user.Service, log *slog.Logger) CallbackHandler {
	return func(c telebot.Context) error {
		if c == nil || userService == nil {
			return nil
		}

		sender := c.Sender()
		if sender == nil {
			return respondCallback(c, "User not found", true)
		}

		data := ""
		if cb := c.Callback(); cb != nil {
			data = cb.Data
		}

		lang := strings.TrimPrefix(data, settingsLanguageDataPrefix)
		if lang == "" {
			return respondCallback(c, "Unknown language option", true)
		}

		ctx := context.Background()
		settings, err := userService.GetSettings(ctx, sender.ID)
		switch {
		case err == nil:
		case errors.Is(err, sql.ErrNoRows):
			settings = defaultUserSettings()
		default:
			if log != nil {
				log.Error("set language: failed to load settings", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return respondCallback(c, "Unable to update settings", true)
		}

		settings.Language = lang
		if err := userService.UpdateSettings(ctx, sender.ID, settings); err != nil {
			if log != nil {
				log.Error("set language: failed to save settings", slog.Int64("telegram_id", sender.ID), slog.Any("error", err))
			}
			return respondCallback(c, "Unable to update settings", true)
		}

		return respondCallback(c, fmt.Sprintf("Language set to %s", strings.ToUpper(lang)), false)
	}
}

func buildSettingsKeyboard(_ *keyboard.Builder, settings *domain.UserSettings) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}

	toggleText := boolLabel(settings.NotificationsEnabled, "Disable notifications", "Enable notifications")
	markup.InlineKeyboard = [][]telebot.InlineButton{
		{
			{
				Text: toggleText,
				Data: settingsToggleNotificationsData,
			},
		},
		{
			{
				Text: "English 🇺🇸",
				Data: settingsLanguageDataPrefix + "en",
			},
			{
				Text: "Русский 🇷🇺",
				Data: settingsLanguageDataPrefix + "ru",
			},
		},
	}

	return markup
}

func boolLabel(value bool, trueLabel, falseLabel string) string {
	if value {
		return trueLabel
	}
	return falseLabel
}

func respondCallback(c telebot.Context, text string, alert bool) error {
	if c == nil {
		return nil
	}
	return c.Respond(&telebot.CallbackResponse{
		Text:      text,
		ShowAlert: alert,
	})
}

func defaultUserSettings() *domain.UserSettings {
	return &domain.UserSettings{
		NotificationsEnabled: true,
		Language:             "en",
		Timezone:             "UTC",
	}
}
