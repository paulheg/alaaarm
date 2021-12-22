package telegram

import (
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		var err error

		text := "Your alerts:\n"
		userAlerts, err := t.repository.GetUserAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(userAlerts) > 0 {
			for _, alert := range userAlerts {
				receiver, _ := t.repository.GetAlertReceiver(alert)

				text += emoji.Sprintf("- :bell:<i>%s</i>, :moai:%v \n", alert.Name, len(receiver))
			}
		} else {
			text += "You have not created an alert.\nSend /create to do so."
		}

		text += "\n\nSubscribed Alerts:\n"

		subscribedAlerts, err := t.repository.GetUserSubscribedAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(subscribedAlerts) > 0 {
			for _, alert := range subscribedAlerts {
				text += emoji.Sprintf("-:bell: %s\n", alert.Name)
			}
		} else {
			text += "You have not subscribed to any alerts yet."
		}

		msg := t.escapedHTMLMessage(u.ChatID, text)
		_, err = t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	}))
}
