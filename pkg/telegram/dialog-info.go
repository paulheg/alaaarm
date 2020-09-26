package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		var err error

		if u.Update.Message.Command() != "info" {
			return dialog.NoMatch, nil
		}

		msg := tgbotapi.NewMessage(u.ChatID, "")

		msg.Text = "Your alerts:\n"
		userAlerts, err := t.data.GetUserAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(userAlerts) > 0 {
			for _, alert := range userAlerts {
				msg.Text += fmt.Sprintf("- %s\n", alert.Name)
			}
			msg.ParseMode = tgbotapi.ModeMarkdown
		} else {
			msg.Text += "You have not created an alert.\nSend /create to do so."
		}

		msg.Text += "\n\nSubscribed Alerts:\n"

		subscribedAlerts, err := t.data.GetUserSubscribedAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(subscribedAlerts) > 0 {
			for _, alert := range subscribedAlerts {
				msg.Text += fmt.Sprintf("- %s\n", alert.Name)
			}
		} else {
			msg.Text += "You have not subscribed to any alerts jet."
		}

		t.bot.Send(msg)
		return dialog.Success, nil
	}))
}
