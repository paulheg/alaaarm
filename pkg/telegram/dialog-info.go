package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		var err error

		msg := tgbotapi.NewMessage(u.ChatID, "")

		msg.Text = "Your alerts:\n"
		userAlerts, err := t.repository.GetUserAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(userAlerts) > 0 {
			for _, alert := range userAlerts {
				msg.Text += emoji.Sprintf("-:bell: %s\n", alert.Name)
			}
			msg.ParseMode = tgbotapi.ModeMarkdown
		} else {
			msg.Text += "You have not created an alert.\nSend /create to do so."
		}

		msg.Text += "\n\nSubscribed Alerts:\n"

		subscribedAlerts, err := t.repository.GetUserSubscribedAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(subscribedAlerts) > 0 {
			for _, alert := range subscribedAlerts {
				msg.Text += emoji.Sprintf("-:bell: %s\n", alert.Name)
			}
		} else {
			msg.Text += "You have not subscribed to any alerts yet."
		}

		_, err = t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	}))
}
