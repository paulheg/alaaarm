package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func userFriendlyAlertIdentifier(alert models.Alert) string {
	return fmt.Sprintf("%d %s", alert.ID, alert.Name)
}

func makeAlertReplyKeyboard(alerts []models.Alert, buttonsPerRow int) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([][]tgbotapi.KeyboardButton, 0)
	var row []tgbotapi.KeyboardButton

	for i, alert := range alerts {

		// if buttonsPerRow is reached
		if i%buttonsPerRow == 0 {
			// in the first iteration i = 0
			// there are no buttons in the row, therefore we dont need to add them
			if i > 0 {
				buttons = append(buttons, row)
			}

			// create a new row
			row = make([]tgbotapi.KeyboardButton, 0)
		}

		// add a new button to the row
		row = append(row,
			tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)),
		)
	}

	// add the remaining buttons to a last row
	buttons = append(buttons, row)

	return tgbotapi.NewReplyKeyboard(buttons...)
}

func (t *Telegram) newAlertSelectionDialog(getAlerts func(u Update) ([]models.Alert, error), onEmptyMessageKey string) *dialog.Dialog {

	const buttonsPerRow = 2

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, err := getAlerts(u)
		if err != nil {
			return dialog.Reset, err
		}

		if len(alerts) == 0 {
			err := t.sendMessage(u, onEmptyMessageKey)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		ctx.Set(ALERTS_CONTEXT_KEY, alerts)

		// build keyboardMarkup
		keyboardMarkup := makeAlertReplyKeyboard(alerts, buttonsPerRow)

		err = t.sendMessageWithReplyMarkup(u, "select_alert", keyboardMarkup)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	}))
}

func (t *Telegram) newSelectSubscribedAlertDialog() *dialog.Dialog {

	return t.newAlertSelectionDialog(
		func(u Update) ([]models.Alert, error) {
			return t.repository.GetUserSubscribedAlerts(u.User.ID)
		},
		"not_subscribed_to_alerts",
	).Append(t.alertDetermination())
}

// Show a select alert dialog (created alerts).
// Places the Alert into the context at ALERT_SELECTION_CONTEXT_KEY
func (t *Telegram) newSelectAlertDialog() *dialog.Dialog {
	return t.newAlertSelectionDialog(
		func(u Update) ([]models.Alert, error) {
			return t.repository.GetUserAlerts(u.User.ID)
		},
		"no_created_alerts",
	).Append(t.alertDetermination())
}

func (t *Telegram) alertDetermination() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, ok := ctx.Value(ALERTS_CONTEXT_KEY).([]models.Alert)
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
			err := t.sendMessage(u, "selected_alert_not_found")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

		alert, err := t.repository.GetAlert(alert.ID)
		if err != nil {
			return dialog.Reset, err
		}

		ctx.Set(ALERT_CONTEXT_KEY, alert)

		err = t.sendCloseKeyboardMessage(u, "alert_was_selected", alert.Name)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Next, nil
	}))
}
