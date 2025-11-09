package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/i18n"
	"github.com/Proton-105/himera-bot/internal/state"
)

const (
	welcomeMessageKey       = "welcome_message"
	welcomeBackMessageKey   = "welcome_back_message"
	internalErrorMessageKey = "internal_error"

	defaultWelcomeMessage       = "Welcome! Let's get you set up."
	defaultWelcomeBackMessage   = "Welcome back!"
	defaultInternalErrorMessage = "An internal error occurred. Please try again later."
)

// NewStartHandler registers the /start command handler that manages user state via the FSM.
func NewStartHandler(fsm state.StateMachine, log *slog.Logger, i18nManager *i18n.Manager) Handler {
	if log == nil {
		log = slog.Default()
	}

	return func(c telebot.Context) error {
		if c == nil || c.Sender() == nil {
			log.Warn("start handler invoked without sender")
			return nil
		}

		if fsm == nil {
			log.Error("state machine is not configured for start handler")
			return c.Send(defaultInternalErrorMessage)
		}

		translator := translatorFor(i18nManager, c.Sender().LanguageCode)
		mainMenu := keyboard.MainMenu(translator)

		ctx := context.Background()
		userID := c.Sender().ID

		_, err := fsm.GetState(ctx, userID)
		switch {
		case err == nil:
			return sendWithMainMenu(c, mainMenu, translator, welcomeBackMessageKey, defaultWelcomeBackMessage)
		case errors.Is(err, state.ErrStateNotFound):
			if setErr := fsm.SetState(ctx, userID, state.StateIdle, nil); setErr != nil {
				log.Error("failed to set initial user state", slog.Int64("telegram_id", userID), slog.Any("error", setErr))
				return c.Send(localizedMessage(translator, internalErrorMessageKey, defaultInternalErrorMessage))
			}
			return sendWithMainMenu(c, mainMenu, translator, welcomeMessageKey, defaultWelcomeMessage)
		default:
			log.Error("failed to fetch user state", slog.Int64("telegram_id", userID), slog.Any("error", err))
			return c.Send(localizedMessage(translator, internalErrorMessageKey, defaultInternalErrorMessage))
		}
	}
}

func translatorFor(manager *i18n.Manager, lang string) i18n.Translator {
	if manager == nil {
		return nil
	}
	return manager.Translator(lang)
}

func localizedMessage(t i18n.Translator, key, fallback string) string {
	if t == nil {
		return fallback
	}

	value := strings.TrimSpace(t.T(key))
	if value == "" || value == key {
		return fallback
	}

	return value
}

func sendWithMainMenu(c telebot.Context, markup *telebot.ReplyMarkup, t i18n.Translator, key, fallback string) error {
	message := localizedMessage(t, key, fallback)
	if markup != nil {
		return c.Send(message, markup)
	}
	return c.Send(message)
}
