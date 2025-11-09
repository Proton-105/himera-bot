package bot

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"fmt"
	"log/slog"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	errors "github.com/Proton-105/himera-bot/internal/errors"
	"github.com/Proton-105/himera-bot/internal/idempotency"
	"github.com/Proton-105/himera-bot/internal/middleware"
	"github.com/Proton-105/himera-bot/internal/repository"
	"github.com/Proton-105/himera-bot/internal/state"
	"github.com/Proton-105/himera-bot/internal/user"
	"github.com/Proton-105/himera-bot/pkg/config"
)

const (
	CommandProfile  = "/profile"
	CommandSettings = "/settings"
)

// Bot wraps telebot.Bot with application dependencies required for handling updates.
type Bot struct {
	telebot            *telebot.Bot
	db                 *sql.DB
	log                *slog.Logger
	cfg                config.Config
	fsm                state.StateMachine
	rateLimitMw        *middleware.RateLimitMiddleware
	router             *Router
	dispatcher         *Dispatcher
	keyboard           *keyboard.Builder
	errHandler         *errors.Handler
	idempotencyManager idempotency.Manager
}

// New builds a telegram bot instance configured according to the application settings.
func New(
	cfg config.Config,
	log *slog.Logger,
	db *sql.DB,
	fsm state.StateMachine,
	idempotencyManager idempotency.Manager,
	rateLimitMw *middleware.RateLimitMiddleware,
	userRepo repository.UserRepository,
	userService *user.Service,
) (*Bot, error) {
	settings := telebot.Settings{
		Token: cfg.Bot.Token,
	}

	if cfg.Bot.Mode == "webhook" {
		settings.Poller = &telebot.Webhook{
			Listen: cfg.Server.Port,
		}
	} else {
		settings.Poller = &telebot.LongPoller{
			Timeout: cfg.Bot.Timeout,
		}
	}

	tb, err := telebot.NewBot(settings)
	if err != nil {
		return nil, fmt.Errorf("initialize telebot: %w", err)
	}

	kb := keyboard.NewBuilder(log)
	dispatcher := NewDispatcher(fsm, log)
	router := NewRouter(dispatcher, log)
	errHandler := errors.NewHandler(log, cfg.Sentry.Enabled)

	b := &Bot{
		telebot:            tb,
		db:                 db,
		log:                log,
		cfg:                cfg,
		fsm:                fsm,
		rateLimitMw:        rateLimitMw,
		router:             router,
		dispatcher:         dispatcher,
		keyboard:           kb,
		errHandler:         errHandler,
		idempotencyManager: idempotencyManager,
	}

	b.setupRouter(userRepo, userService, log)

	if b.rateLimitMw != nil {
		b.telebot.Use(b.rateLimitMw.Handle)
	}

	b.registerTelebotHandlers()

	return b, nil
}

// Start runs the telegram bot event loop.
func (b *Bot) Start() {
	if b.telebot != nil {
		b.telebot.Start()
	}
}

// Stop gracefully stops the telegram bot.
func (b *Bot) Stop() {
	if b.telebot == nil {
		return
	}

	if b.log != nil {
		b.log.Info("stopping telegram bot...")
	}

	b.telebot.Stop()
}

// Telebot exposes the underlying telebot.Bot instance for integrations such as health checks.
func (b *Bot) Telebot() *telebot.Bot {
	return b.telebot
}

func (b *Bot) setupRouter(userRepo repository.UserRepository, userService *user.Service, log *slog.Logger) {
	if b.router == nil {
		return
	}

	b.router.Use(RecoveryMiddleware(b.log, b.errHandler))
	b.router.Use(middleware.Idempotency(b.idempotencyManager, b.log))
	b.router.Use(ErrorHandlingMiddleware(b.errHandler))
	b.router.Use(LoggingMiddleware(b.log))
	b.router.Use(AuthMiddleware(userRepo, b.log))
	b.router.Use(LastActiveMiddleware(userService))
	b.router.Use(middleware.Metrics)

	b.router.RegisterCommand(CommandStart, newStartCommandHandler(b.fsm, b.log))
	b.router.RegisterCommand(CommandCancel, handlers.NewCancelHandler(b.fsm, b.keyboard, b.log))

	if userService == nil {
		return
	}

	profileHandler := handlers.NewProfileHandler(userService, log)
	b.router.RegisterCommand(CommandProfile, profileHandler)

	settingsHandler := handlers.NewSettingsHandler(userService, b.keyboard, log)
	b.router.RegisterCommand(CommandSettings, settingsHandler)

	b.router.RegisterCallback("settings_toggle_notifications", handlers.HandleToggleNotifications(userService, log))
	b.router.RegisterCallback("settings_set_language_", handlers.HandleSetLanguage(userService, log))
}

func (b *Bot) registerTelebotHandlers() {
	if b.telebot == nil || b.router == nil {
		return
	}

	b.telebot.Handle(telebot.OnText, b.router.Route)
	b.telebot.Handle(telebot.OnCallback, b.router.Route)
}

func newStartCommandHandler(fsm state.StateMachine, log *slog.Logger) handlers.Handler {
	return func(c telebot.Context) error {
		if c == nil || c.Sender() == nil {
			if log != nil {
				log.Warn("start handler invoked without sender")
			}
			return nil
		}

		if fsm == nil {
			if log != nil {
				log.Error("state machine not configured for start handler")
			}
			return c.Send("An internal error occurred. Please try again later.")
		}

		ctx := context.Background()
		userID := c.Sender().ID

		_, err := fsm.GetState(ctx, userID)
		switch {
		case err == nil:
			return c.Send("Welcome back!")
		case stdErrors.Is(err, state.ErrStateNotFound):
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
}
