package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newDeleteDialog() *dialog.Dialog {
	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")

			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.Text = emoji.Sprintf(`:warning: *Do you really want to delete the :bell: %s alert?*
No one will receive any notifications from this alert anymore.
HTTP Requests using the token of this alert wont result in a notification.

:warning: This process cannot be reversed.
Do you want to delete the alert?`, alert.Name)
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Next, nil
		})).Append(t.newYesNoDialog(
		// On Yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			var err error
			msg := tgbotapi.NewMessage(u.ChatID, "")

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			err = t.repository.DeleteAlert(alert)
			if err != nil {
				return dialog.Reset, err
			}

			msg.Text = emoji.Sprint(":check_mark_button: Alert was deleted.")
			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// On No
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprint(":cross_mark: The alert was not deleted.")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
	))
}
