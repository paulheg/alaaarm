package telegram

// Configuration of the telegram bot
type Configuration struct {
	APIKey string `env:"TELEGRAM_API_KEY"`
}
