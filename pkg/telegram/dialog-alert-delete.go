package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newDeleteDialog() *dialog.Dialog {
	const AlertKey = "alert"

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")

			msg.Text = emoji.Sprintf(`:warning: Do you really want to delete the %s alert.
No one will receive any notifications from this alert anymore.
HTTP Requests using the token of this alert wont result in a notification.

:warning: This process cannot be reversed.
Do you want to delete the alert?`, alert.Name)
			t.bot.Send(msg)

			return dialog.Next, nil
		})).Append(t.newYesNoDialog(
		// On Yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			var err error
			msg := tgbotapi.NewMessage(u.ChatID, "")

			alert, ok := ctx.Value(AlertKey).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			err = t.repository.DeleteAlert(alert)
			if err != nil {
				return dialog.Reset, err
			}

			msg.Text = "Alert was deleted."
			t.bot.Send(msg)
			return dialog.Success, nil
		},
		// On No
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "The alert was not deleted.")
			t.bot.Send(msg)
			return dialog.Success, nil
		},
	))
}
