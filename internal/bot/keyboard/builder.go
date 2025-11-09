package keyboard

import (
	"log/slog"

	telebot "gopkg.in/telebot.v3"
)

// Builder creates inline keyboards based on the user's state.
type Builder struct {
	log *slog.Logger
}

// NewBuilder returns a new Builder instance.
func NewBuilder(log *slog.Logger) *Builder {
	return &Builder{log: log}
}

// MainMenu builds the idle state menu.
func (b *Builder) MainMenu() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.InlineKeyboard = [][]telebot.InlineButton{
		{
			{
				Text: "Buy 💰",
				Data: "buy",
			},
		},
		{
			{
				Text: "Sell 📉",
				Data: "sell",
			},
		},
		{
			{
				Text: "Portfolio 📊",
				Data: "portfolio",
			},
		},
	}
	return markup
}

// AmountButtons builds quick amount selection buttons.
func (b *Builder) AmountButtons() *telebot.ReplyMarkup {
	values := []string{"50", "100", "500", "1000"}
	row := make([]telebot.InlineButton, 0, len(values))
	for _, value := range values {
		row = append(row, telebot.InlineButton{
			Text: value,
			Data: "amount_" + value,
		})
	}

	markup := &telebot.ReplyMarkup{}
	markup.InlineKeyboard = [][]telebot.InlineButton{row}
	return markup
}

// ConfirmButtons builds confirmation buttons for buy or sell actions.
func (b *Builder) ConfirmButtons(action string) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.InlineKeyboard = [][]telebot.InlineButton{
		{
			{
				Text: "Confirm ✅",
				Data: action + "_confirm",
			},
			{
				Text: "Cancel ❌",
				Data: action + "_cancel",
			},
		},
	}
	return markup
}

// CancelButton builds a single cancel button.
func (b *Builder) CancelButton() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.InlineKeyboard = [][]telebot.InlineButton{
		{
			{
				Text: "Cancel ❌",
				Data: "cancel",
			},
		},
	}
	return markup
}
