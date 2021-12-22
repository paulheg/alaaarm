package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

const (
	ALERT_SELECTION_CONTEXT_KEY = "alert"
)

func userFriendlyAlertIdentifier(alert models.Alert) string {
	return fmt.Sprintf("%d %s", alert.ID, alert.Name)
}

func (t *Telegram) newAlertSelectionDialog(getAlerts func(u Update) ([]models.Alert, error), onEmptyMessage string) *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, err := getAlerts(u)
		if err != nil {
			return dialog.Reset, err
		}

		if len(alerts) == 0 {
			_, err := t.bot.Send(t.escapedHTMLMessage(u.ChatID, onEmptyMessage))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		ctx.Set("alerts", alerts)

		// build keyboard
		buttons := make([]tgbotapi.KeyboardButton, 0)

		for _, alert := range alerts {
			buttons = append(buttons,
				tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)),
			)
		}

		msg := t.escapedHTMLMessage(u.ChatID, "Select the alert")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)

		_, err = t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	}))
}

func (t *Telegram) newSelectSubscribedAlertDialog() *dialog.Dialog {
	return t.newAlertSelectionDialog(func(u Update) ([]models.Alert, error) {
		return t.repository.GetUserSubscribedAlerts(u.User.ID)
	}, "You are not subscribed to any alerts yet.").Append(t.alertDetermination())
}

// Show a select alert dialog (created alerts).
// Places the Alert into the context at ALERT_SELECTION_CONTEXT_KEY
func (t *Telegram) newSelectAlertDialog() *dialog.Dialog {
	return t.newAlertSelectionDialog(func(u Update) ([]models.Alert, error) {
		return t.repository.GetUserAlerts(u.User.ID)
	}, "You have not created any alerts yet.").Append(t.alertDetermination())
}

func (t *Telegram) alertDetermination() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, ok := ctx.Value("alerts").([]models.Alert)
		if !ok {
			return dialog.Reset, errContextDataMissing
		}

		var alert models.Alert
		foundAlert := false

		alertIdentifier := u.Text
		for _, alert = range alerts {
			if userFriendlyAlertIdentifier(alert) == alertIdentifier {
				foundAlert = true
				break
			}
		}

		if !foundAlert {
			msg := t.escapedHTMLMessage(u.ChatID, "Could not find the alert you selected")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

		alert, err := t.repository.GetAlert(alert.ID)
		if err != nil {
			return dialog.Reset, err
		}

		ctx.Set("alert", alert)

		msg := t.escapedHTMLMessage(u.ChatID, "You selected the <b>%s</b> alert.", alert.Name)

		// reset keyboard
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		}
		_, err = t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Next, nil
	}))
}
