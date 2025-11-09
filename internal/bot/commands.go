package bot

// Command constants for Telegram bot commands.
const (
	CommandStart     = "/start"
	CommandBuy       = "/buy"
	CommandSell      = "/sell"
	CommandPortfolio = "/portfolio"
	CommandCancel    = "/cancel"
	CommandHelp      = "/help"
)

// Callback prefix constants for inline button interactions.
const (
	CallbackBuyConfirm  = "buy_confirm"
	CallbackBuyCancel   = "buy_cancel"
	CallbackSellConfirm = "sell_confirm"
	CallbackSellCancel  = "sell_cancel"
)
