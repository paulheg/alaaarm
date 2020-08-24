package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			update := i.(Update)
			var err error

			if update.Update.Message.Command() != "info" {
				return dialog.ErrNoMatch
			}

			msg := tgbotapi.NewMessage(update.ChatID, "")

			msg.Text = "Your alerts:\n"
			userAlerts, err := t.data.GetUserAlerts(update.User.ID)
			if err == nil {
				if len(userAlerts) > 0 {
					for _, alert := range userAlerts {
						msg.Text += fmt.Sprintf("- %s\n", alert.Name)
					}
					msg.ParseMode = tgbotapi.ModeMarkdown
				} else {
					msg.Text += "You have not created an alert.\nSend /create to do so."
				}
			} else {
				msg.Text = "There was an error while gathering your alerts."
				return dialog.ErrReset
			}

			msg.Text += "\n\nSubscribed Alerts:\n"

			subscribedAlerts, err := t.data.GetUserSubscribedAlerts(update.User.ID)
			if err == nil {
				if len(subscribedAlerts) > 0 {
					for _, alert := range subscribedAlerts {
						msg.Text += fmt.Sprintf("- %s\n", alert.Name)
					}
				} else {
					msg.Text += "You have not subscribed to any alerts jet."
				}
			} else {
				msg.Text = "There was an error while gathering your alerts."
				return dialog.ErrReset
			}

			t.bot.Send(msg)
			return err
		})

}
