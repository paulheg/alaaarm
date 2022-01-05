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

			err = t.sendMessage(u, "alert_delete_are_you_shure", alert.Name, subscribers)
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

			err = t.sendMessage(u, "alert_deleted")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// On No
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			err := t.sendMessage(u, "alert_not_deleted")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
	))
}
