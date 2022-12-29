package telegram

import (
	"database/sql"
	"strings"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

// newStartDialog
// dialog when user first open the bot
// invite id is sent as a parameter
func (t *Telegram) newStartDialog() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		invitationKey := u.Update.Message.CommandArguments()
		invitationKey = strings.TrimSpace(invitationKey)

		if len(invitationKey) == 0 {
			// normal start command

			err := t.sendMessage(u, "welcome_message")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		// check if arugment passed to start argument is an invitation key
		invite, err := t.repository.GetInviteByToken(invitationKey)
		if err == sql.ErrNoRows {
			err := t.sendMessage(u, "invitation_does_not_exist")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		} else if err != nil {
			return dialog.Reset, err
		}

		if invite.Alert.Owner.ID == u.User.ID {
			t.sendMessage(u, "inivite_to_own_alert")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		// safe to context
		ctx.Set(INVITE_CONTEXT_KEY, invite)

		err = t.sendMessage(u, "join_alert",
			invite.Alert.Name,
			invite.Alert.Description,
			invite.Alert.Owner.TelegramUserLink())

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

			err := t.repository.AddUserToAlert(*invite.Alert, u.User)
			if err != nil {
				return dialog.Reset, err
			}

			err = t.sendMessage(u, "succesful_join", invite.Alert.Name)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			err := t.sendMessage(u, "didnt_join")
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		},
	))

}
