package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/messages"
	"github.com/paulheg/alaaarm/pkg/models"
)

// Update represents the struct passed to the dialog handler
type Update struct {
	Update     tgbotapi.Update
	User       models.User
	Text       string
	ChatID     int64
	Dictionary messages.Dictionary
	Language   string
}
