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

			_, err := t.bot.Send(t.escapedHTMLMessage(u.ChatID, `Welcome to the :bell: Alaaarm bot.
With this bot you can create and receive alerts.

To create an alert use /create`))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		// check if arugment passed to start argument is an invitation key
		invite, err := t.repository.GetInviteByToken(invitationKey)
		if err == sql.ErrNoRows {
			_, err = t.bot.Send(
				t.escapedHTMLMessage(u.ChatID, ":cross_mark: The invitation does no longer exist."))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		} else if err != nil {
			return dialog.Reset, err
		}

		if invite.Alert.Owner.ID == u.User.ID {
			_, err = t.bot.Send(
				t.escapedHTMLMessage(u.ChatID, ":warning: You are the owner of this alert, you already will be notified."))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		}

		// safe to context
		ctx.Set("invite", invite)

		msg := t.escapedHTMLMessage(u.ChatID, `Do you want to join the following :bell: alert?
<b>%s</b>
%s
Of <a href="%s">Owner</a>`,
			invite.Alert.Name,
			invite.Alert.Description,
			invite.Alert.Owner.TelegramUserLink())

		_, err = t.bot.Send(msg)
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

			err := t.repository.AddUserToAlert(*invite.Alert, u.User)
			if err != nil {
				return dialog.Reset, err
			}

			msg := t.escapedHTMLMessage(u.ChatID,
				":check_mark_button: You successfully joined the %s alert. You will be notified on the next alert.",
				invite.Alert.Name)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := t.escapedHTMLMessage(u.ChatID, ":cross_mark: You did not join.")

			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		},
	))

}
