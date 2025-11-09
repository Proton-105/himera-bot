package handlers

import (
	telebot "gopkg.in/telebot.v3"
)

// Handler processes bot commands.
type Handler func(c telebot.Context) error

// CallbackHandler processes inline callback events.
type CallbackHandler func(c telebot.Context) error

// Middleware wraps handlers with additional behavior.
type Middleware func(Handler) Handler

// HandlerFunc adapts ordinary functions to the Handler interface.
type HandlerFunc func(c telebot.Context) error

// Handle executes the underlying function.
func (h HandlerFunc) Handle(c telebot.Context) error {
	return h(c)
}
