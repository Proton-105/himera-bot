package keyboard

import (
	telebot "gopkg.in/telebot.v3"
)

// InlineButton represents a lightweight inline keyboard button definition used by the builder.
type InlineButton struct {
	Text   string
	Unique string // Identifier that differentiates callback handlers.
	Data   string // Payload that will be encoded into callback data.
}

// InlineKeyboardBuilder accumulates rows of InlineButton definitions before rendering telebot markup.
type InlineKeyboardBuilder struct {
	markup *telebot.ReplyMarkup
	rows   [][]InlineButton
}

// NewInlineKeyboard creates a builder instance backed by inline reply markup.
func NewInlineKeyboard() *InlineKeyboardBuilder {
	return &InlineKeyboardBuilder{
		markup: &telebot.ReplyMarkup{InlineKeyboard: make([][]telebot.InlineButton, 0)},
		rows:   make([][]InlineButton, 0),
	}
}

// AddRow appends a new row made of custom InlineButton definitions.
func (b *InlineKeyboardBuilder) AddRow(buttons ...InlineButton) *InlineKeyboardBuilder {
	if len(buttons) == 0 {
		return b
	}

	row := make([]InlineButton, len(buttons))
	copy(row, buttons)
	b.rows = append(b.rows, row)
	return b
}

// Build finalizes inline markup using the provided encoder to produce callback data strings.
func (b *InlineKeyboardBuilder) Build(encoder func(unique, data string) string) *telebot.ReplyMarkup {
	if encoder == nil {
		encoder = func(unique, data string) string {
			if data != "" {
				return data
			}
			return unique
		}
	}

	if b.markup == nil {
		b.markup = &telebot.ReplyMarkup{}
	}

	inlineKeyboard := make([][]telebot.InlineButton, len(b.rows))
	for i, row := range b.rows {
		inlineKeyboard[i] = make([]telebot.InlineButton, len(row))
		for j, btn := range row {
			inlineKeyboard[i][j] = telebot.InlineButton{
				Text:   btn.Text,
				Unique: btn.Unique,
				Data:   encoder(btn.Unique, btn.Data),
			}
		}
	}

	b.markup.InlineKeyboard = inlineKeyboard
	return b.markup
}
