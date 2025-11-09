package keyboard_test

import (
	"strings"
	"testing"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/testutil"
)

func TestInlineKeyboardBuilder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		builder := keyboard.NewInlineKeyboard()
		builder.AddRow(
			keyboard.InlineButton{Text: "Prev", Unique: "nav", Data: "1"},
			keyboard.InlineButton{Text: "Next", Unique: "nav", Data: "2"},
		).AddRow(
			keyboard.InlineButton{Text: "Confirm", Unique: "confirm", Data: "ok"},
		)

		markup, err := builder.Build()
		testutil.AssertNoError(t, err)

		if markup == nil {
			t.Fatal("expected markup, got nil")
		}

		if len(markup.InlineKeyboard) == 0 {
			t.Fatal("expected inline keyboard rows")
		}

		testutil.AssertEqual(t, 2, len(markup.InlineKeyboard))
		testutil.AssertEqual(t, 2, len(markup.InlineKeyboard[0]))
		testutil.AssertEqual(t, 1, len(markup.InlineKeyboard[1]))
		testutil.AssertEqual(t, "nav:2", markup.InlineKeyboard[0][1].Data)
	})

	t.Run("callback data overflow", func(t *testing.T) {
		builder := keyboard.NewInlineKeyboard()
		builder.AddRow(keyboard.InlineButton{
			Text:   "Too big",
			Unique: "overflow",
			Data:   strings.Repeat("x", keyboard.CallbackDataLimitBytes),
		})

		_, err := builder.Build()
		testutil.AssertError(t, err)
	})
}
