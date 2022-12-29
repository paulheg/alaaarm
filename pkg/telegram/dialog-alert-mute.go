package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newMuteAlertDialog() *dialog.Dialog {
	return t.newSelectAlertDialog().Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alert, ok := ctx.Value(ALERT_CONTEXT_KEY).(models.Alert)
		if !ok {
			return dialog.Reset, errContextDataMissing
		}

		t.repository.SetMuteAlert(alert, !alert.NotifyOwner)
		if alert.NotifyOwner {
			t.sendMessage(u, "now_muted")
		} else {
			t.sendMessage(u, "now_unmuted")
		}

		return dialog.Success, nil
	}))
}
