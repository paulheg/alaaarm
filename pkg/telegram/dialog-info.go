package telegram

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			update := i.(tgbotapi.Update)
			var err error

			if update.Message != nil {

				if update.Message.IsCommand() && update.Message.Command() == "info" {
					var response string
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

					var userAlerts []models.Alert

					exists, user, _ := t.data.GetUserTelegram(update.Message.Chat.ID)
					if !exists {
						t.data.CreateUser(models.User{
							Username:   update.Message.Chat.UserName,
							TelegramID: update.Message.Chat.ID,
						})
					}

					userAlerts, err := t.data.GetUserAlerts(user.ID)
					if err == nil {
						if len(userAlerts) > 0 {
							response = "Your alerts:\n"
							var rows []tgbotapi.InlineKeyboardButton
							for _, alert := range userAlerts {

								row := tgbotapi.NewInlineKeyboardRow(
									tgbotapi.NewInlineKeyboardButtonData(
										alert.Name,
										strconv.Itoa(int(alert.ID)),
									),
								)

								rows = append(rows, row...)
							}
							msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows)
						} else {
							response += "You have not created an alert.\nSend /create to do so."
						}
					}

					msg.Text = response
					msg.ParseMode = tgbotapi.ModeMarkdown
					t.bot.Send(msg)
				} else {
					err = dialog.ErrNoMatch
				}
			} else {
				err = dialog.ErrNoMatch
			}

			return err
		},
	)

}
