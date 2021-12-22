package telegram

import (
	"database/sql"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newInviteDeleteDiaolg() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			invite, err := t.repository.GetInviteByAlertID(alert.ID)
			if err == sql.ErrNoRows {
				_, err := t.bot.Send(
					t.escapedHTMLMessage(u.ChatID, "There was no invite to delete for the %s alert.", alert.Name))
				if err != nil {
					return dialog.Reset, err
				}

				return dialog.Reset, nil
			} else if err != nil {
				return dialog.Reset, err
			}

			ctx.Set("invite", invite)

			_, err = t.bot.Send(
				t.escapedHTMLMessage(u.ChatID, "Do you want to delete the invite for the %s alert?", alert.Name))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Next, nil
		})).Append(t.newYesNoDialog(
		// on yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			invite, ok := ctx.Value("invite").(models.Invite)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			err := t.repository.DeleteInvite(invite.ID)
			if err != nil {
				return dialog.Reset, err
			}

			_, err = t.bot.Send(
				t.escapedHTMLMessage(u.ChatID, ":check_mark_button: The invite was successfuly deleted."))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			_, err := t.bot.Send(
				t.escapedHTMLMessage(u.ChatID, ":cross_mark: The invite was not deleted."))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
	))

}
