package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newInfoDialog() *dialog.Dialog {

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		var err error
		var text string

		if s, err := t.lookupText(u, "alerts_heading"); err != nil {
			return dialog.Reset, err
		} else {
			text += s + "\n"
		}

		userAlerts, err := t.repository.GetUserAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(userAlerts) > 0 {
			for _, alert := range userAlerts {
				receiver, _ := t.repository.GetAlertReceiver(alert)

				text += t.emojify("- :bell:<i>%s</i>, :moai:%v \n", alert.Name, len(receiver))
			}
		} else {

			if s, err := t.lookupText(u, "no_created_alerts"); err != nil {
				return dialog.Reset, err
			} else {
				text += s
			}
		}

		if s, err := t.lookupText(u, "subscribed_alerts_heading"); err != nil {
			return dialog.Reset, err
		} else {
			text += "\n\n" + s + "\n"
		}

		subscribedAlerts, err := t.repository.GetUserSubscribedAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		if len(subscribedAlerts) > 0 {
			for _, alert := range subscribedAlerts {
				text += t.emojify("-:bell: %s\n", alert.Name)
			}
		} else {

			if s, err := t.lookupText(u, "not_subscribed_to_alerts"); err != nil {
				return dialog.Reset, err
			} else {
				text += s
			}
		}

		msg := t.escapedHTMLMessage(u.ChatID, text)
		_, err = t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	}))
}
