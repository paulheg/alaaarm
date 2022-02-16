package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertChangeTokenDialog() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			updatedAlert, err := t.repository.UpdateAlertToken(alert)
			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(updatedAlert, "Hello World")

			err = t.sendMessage(u, "token_changed", updatedAlert.Name,
				triggerURL,
				triggerURL,
			)

			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))

}
