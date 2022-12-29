package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertUnsubscribeDialog() *dialog.Dialog {

	return t.newSelectSubscribedAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_CONTEXT_KEY).(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			if u.User.ID == alert.OwnerID {
				err := t.sendMessage(u, "cant_unsubscribe_from_own_alert")
				if err != nil {
					return dialog.Reset, err
				}

				return dialog.Success, nil
			}

			err := t.repository.RemoveUserFromAlert(alert, u.User)

			if err != nil {
				return dialog.Reset, err
			}

			err = t.sendMessage(u, "succesful_unsubscribe", alert.Name)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))

}
