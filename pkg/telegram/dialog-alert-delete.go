package telegram

import (
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

			subscribers, err := t.repository.GetSubscriberCount(alert)
			if err != nil {
				return dialog.Reset, err
			}

			msg := t.escapedHTMLLookup(u.ChatID, "alert_delete_are_you_shure", alert.Name, subscribers)
			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Next, nil
		})).Append(t.newYesNoDialog(
		// On Yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			var err error

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			err = t.repository.DeleteAlert(alert)
			if err != nil {
				return dialog.Reset, err
			}

			msg := t.escapedHTMLLookup(u.ChatID, "alert_deleted")
			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// On No
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := t.escapedHTMLLookup(u.ChatID, "alert_not_deleted")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
	))
}
