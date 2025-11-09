package keyboard

import (
	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/i18n"
)

// MainMenu builds a localized reply keyboard for the bot main menu.
func MainMenu(t i18n.Translator) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}

	lookup := func(key string) string {
		if t == nil {
			return key
		}
		return t.T(key)
	}

	buyBtn := markup.Text(lookup("main_menu.buy"))
	sellBtn := markup.Text(lookup("main_menu.sell"))
	portfolioBtn := markup.Text(lookup("main_menu.portfolio"))
	balanceBtn := markup.Text(lookup("main_menu.balance"))
	topTokensBtn := markup.Text(lookup("main_menu.top_tokens"))
	historyBtn := markup.Text(lookup("main_menu.history"))
	settingsBtn := markup.Text(lookup("main_menu.settings"))

	markup.Reply(
		markup.Row(buyBtn, sellBtn),
		markup.Row(portfolioBtn, balanceBtn),
		markup.Row(topTokensBtn, historyBtn),
		markup.Row(settingsBtn),
	)

	return markup
}
