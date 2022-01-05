package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertChangeTokenDialog() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			updatedAlert, err := t.repository.UpdateAlertToken(alert)
			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(updatedAlert, "Hello World")

			text, err := u.Dictionary.Lookup("token_changed")
			if err != nil {
				return dialog.Reset, err
			}

			msg := t.escapedHTMLMessage(u.ChatID, text,
				updatedAlert.Name,
				triggerURL,
				triggerURL,
			)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))

}
