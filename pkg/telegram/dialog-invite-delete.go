package telegram

import (
	"database/sql"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newInviteDeleteDiaolg() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			invite, err := t.repository.GetInviteByAlertID(alert.ID)
			if err == sql.ErrNoRows {

				err := t.sendMessage(u, "no_invite_existing", alert.Name)
				if err != nil {
					return dialog.Reset, err
				}

				return dialog.Reset, nil
			} else if err != nil {
				return dialog.Reset, err
			}

			ctx.Set(INVITE_CONTEXT_KEY, invite)

			err = t.sendMessage(u, "delete_invite_question", alert.Name)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Next, nil
		})).Append(t.newYesNoDialog(
		// on yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			invite, ok := ctx.Value(INVITE_CONTEXT_KEY).(models.Invite)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			err := t.repository.DeleteInvite(invite.ID)
			if err != nil {
				return dialog.Reset, err
			}

			err = t.sendMessage(u, "invite_deleted")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			err := t.sendMessage(u, "invite_not_deleted")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
	))

}
