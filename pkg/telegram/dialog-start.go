package telegram

import (
	"database/sql"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

// newStartDialog
// dialog when user first open the bot
// invite id is sent as a parameter
func (t *Telegram) newStartDialog() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")
		invitationKey := u.Update.Message.CommandArguments()
		invitationKey = strings.TrimSpace(invitationKey)

		if len(invitationKey) == 0 {
			// normal start command
			msg.Text = emoji.Sprint(`Welcome to the :bell: Alaaarm bot.
With this bot you can create and receive alerts.

To create an alert use /create`)
			t.bot.Send(msg)
			return dialog.Reset, nil
		}

		// check if arugment passed to start argument is an invitation key
		invite, err := t.repository.GetInviteByToken(invitationKey)
		if err == sql.ErrNoRows {
			msg.Text = emoji.Sprint(":cross_mark: The invitation does no longer exist.")
			t.bot.Send(msg)
			return dialog.Reset, nil
		} else if err != nil {
			return dialog.Reset, err
		}

		if invite.Alert.Owner.ID == u.User.ID {
			msg.Text = emoji.Sprint(":warning: You are the owner of this alert, you already will be notified.")
			t.bot.Send(msg)
			return dialog.Reset, nil
		}

		// safe to context
		ctx.Set("invite", invite)

		joinMessage := `Do you want to join the following alert?
__%s__
_%s_
Of [Owner](%s)`

		msg.Text = fmt.Sprintf(joinMessage,
			invite.Alert.Name,
			invite.Alert.Description,
			invite.Alert.Owner.TelegramUserLink(),
		)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		t.bot.Send(msg)

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

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprintf(":check_mark_button: You successfully joined the %s alert. You will be notified on the next alert.",
				invite.Alert.Name,
			)
			t.bot.Send(msg)
			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprint(":cross_mark: You did not join.")

			t.bot.Send(msg)
			return dialog.Reset, nil
		},
	))

}
