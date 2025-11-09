package bot

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/internal/state"
)

// Router dispatches commands, callbacks, and state-aware updates.
type Router struct {
	mu             sync.RWMutex
	commands       map[string]handlers.Handler
	callbacks      map[string]handlers.CallbackHandler
	dispatcher     *Dispatcher
	defaultHandler handlers.Handler
	middlewares    []handlers.Middleware
	log            *slog.Logger
}

// NewRouter builds a Router with empty registries.
func NewRouter(dispatcher *Dispatcher, log *slog.Logger) *Router {
	if log == nil {
		log = slog.Default()
	}

	return &Router{
		commands:    make(map[string]handlers.Handler),
		callbacks:   make(map[string]handlers.CallbackHandler),
		dispatcher:  dispatcher,
		middlewares: make([]handlers.Middleware, 0),
		log:         log,
	}
}

// RegisterCommand registers a handler for a bot command.
func (r *Router) RegisterCommand(cmd string, h handlers.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd] = h
}

// RegisterCallback registers a handler for callback data prefixes.
func (r *Router) RegisterCallback(prefix string, h handlers.CallbackHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbacks[prefix] = h
}

// Use appends a middleware to the chain.
func (r *Router) Use(mw handlers.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw)
}

// SetDefault sets the fallback handler for unmatched commands or states.
func (r *Router) SetDefault(h handlers.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = h
}

// Route directs the incoming update to the appropriate handler.
func (r *Router) Route(c telebot.Context) error {
	if c == nil {
		return nil
	}

	if callback := c.Callback(); callback != nil {
		return r.handleCallback(c, callback.Data)
	}

	return r.handleMessage(c)
}

func (r *Router) handleCallback(c telebot.Context, data string) error {
	handler := r.findCallbackHandler(data)
	if handler == nil {
		r.log.Info("no callback handler found", "data", data)
		return nil
	}

	exec := handlers.Handler(func(ctx telebot.Context) error {
		return handler(ctx)
	})

	return r.executeHandler(exec, c)
}

func (r *Router) handleMessage(c telebot.Context) error {
	text := c.Text()

	if strings.HasPrefix(text, "/") {
		if handler := r.getCommandHandler(text); handler != nil {
			return r.executeHandler(handler, c)
		}
	}

	stateHandled, err := r.dispatchState(c)
	if err != nil {
		return err
	}
	if stateHandled {
		return nil
	}

	if handler := r.getDefaultHandler(); handler != nil {
		return r.executeHandler(handler, c)
	}

	return nil
}

func (r *Router) dispatchState(c telebot.Context) (bool, error) {
	if r.dispatcher == nil {
		return false, nil
	}

	hasHandler, err := r.stateHandlerExists(c)
	if err != nil {
		return false, err
	}

	if err := r.dispatcher.Dispatch(c); err != nil {
		return false, err
	}

	return hasHandler, nil
}

func (r *Router) stateHandlerExists(c telebot.Context) (bool, error) {
	if r.dispatcher == nil || c == nil || c.Sender() == nil {
		return false, nil
	}

	ctx := context.Background()
	userID := c.Sender().ID

	currentState := state.StateIdle
	userState, err := r.dispatcher.fsm.GetState(ctx, userID)
	if err != nil {
		if !errors.Is(err, state.ErrStateNotFound) {
			return false, err
		}
	} else if userState != nil {
		currentState = userState.CurrentState
	}

	return r.dispatcher.getHandler(currentState) != nil, nil
}

func (r *Router) executeHandler(h handlers.Handler, c telebot.Context) error {
	wrapped := r.applyMiddlewares(h)
	if wrapped == nil {
		return nil
	}
	return wrapped(c)
}

func (r *Router) findCallbackHandler(data string) handlers.CallbackHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for prefix, handler := range r.callbacks {
		if strings.HasPrefix(data, prefix) {
			return handler
		}
	}

	return nil
}

func (r *Router) getCommandHandler(cmd string) handlers.Handler {
	r.mu.RLock()
	handler := r.commands[cmd]
	r.mu.RUnlock()
	return handler
}

func (r *Router) getDefaultHandler() handlers.Handler {
	r.mu.RLock()
	handler := r.defaultHandler
	r.mu.RUnlock()
	return handler
}

// applyMiddlewares wraps the handler with all registered middlewares.
func (r *Router) applyMiddlewares(h handlers.Handler) handlers.Handler {
	if h == nil {
		return nil
	}

	middlewares := r.middlewaresSnapshot()
	wrapped := h
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}

func (r *Router) middlewaresSnapshot() []handlers.Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.middlewares) == 0 {
		return nil
	}

	snapshot := make([]handlers.Middleware, len(r.middlewares))
	copy(snapshot, r.middlewares)
	return snapshot
}
