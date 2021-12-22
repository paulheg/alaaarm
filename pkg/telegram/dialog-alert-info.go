package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertInfoDialog() *dialog.Dialog {

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

			triggerURL := t.webserver.AlertTriggerURL(alert, "Hello World")

			msg := t.escapedHTMLLookup(u.ChatID, "alert_info",
				alert.Name,
				alert.Description,
				subscribers,
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
