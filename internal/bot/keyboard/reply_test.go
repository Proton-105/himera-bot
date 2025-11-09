package keyboard_test

import (
	"testing"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/testutil"
)

func TestMainMenu(t *testing.T) {
	translator := &mockTranslator{
		translations: map[string]string{
			"main_menu.buy":        "Buy",
			"main_menu.sell":       "Sell",
			"main_menu.portfolio":  "Portfolio",
			"main_menu.balance":    "Balance",
			"main_menu.top_tokens": "Top Tokens",
			"main_menu.history":    "History",
			"main_menu.settings":   "Settings",
		},
	}

	markup := keyboard.MainMenu(translator)

	if !markup.ResizeKeyboard {
		t.Fatalf("expected ResizeKeyboard to be true")
	}

	expectedRows := [][]string{
		{"Buy", "Sell"},
		{"Portfolio", "Balance"},
		{"Top Tokens", "History"},
		{"Settings"},
	}

	testutil.AssertEqual(t, len(expectedRows), len(markup.ReplyKeyboard))

	for i, row := range expectedRows {
		testutil.AssertEqual(t, len(row), len(markup.ReplyKeyboard[i]))
		for j, text := range row {
			testutil.AssertEqual(t, text, markup.ReplyKeyboard[i][j].Text)
		}
	}
}
