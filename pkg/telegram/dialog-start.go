package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

// newStartDialog
// dialog when user first open the bot
// invite id is sent as a parameter
func (t *Telegram) newStartDialog() *dialog.Dialog {
	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error
			update := i.(Update)

			if update.Update.Message.Command() != "start" {
				return dialog.ErrNoMatch
			}

			msg := tgbotapi.NewMessage(update.ChatID, "")
			invitationKey := update.Update.Message.CommandArguments()

			if len(invitationKey) > 0 {
				// check if arugment passed to start argument is an invitation key
				exists, invite, err := t.data.GetInvitation(invitationKey)
				if exists && err == nil {

					if invite.Alert.OwnerID == update.User.ID {
						msg.Text = "You are the owner of this alert, you already will be notified."
						t.bot.Send(msg)
						return dialog.ErrReset
					}

					msg.Text = fmt.Sprintf("Do you want to join the %s alert from %s ?",
						invite.Alert.Name,
						invite.Alert.Owner.Username,
					)
					msg.ReplyMarkup = yesNoMenuKeyboard

					// safe to context
					ctx.Set("invite", invite)

				} else {
					msg.Text = "The invitation does no longer exist."
				}
			} else {
				// normal start command
				msg.Text = "Welcome to the Alaaarm bot.\nWith this bot you can create and receive alerts."
				t.bot.Send(msg)
				return dialog.ErrReset
			}

			t.bot.Send(msg)
			return err
		},
	).Chain(func(i interface{}, ctx dialog.ValueStore) error {
		var err error
		update := i.(Update)

		msg := tgbotapi.NewMessage(update.ChatID, "")

		if update.Text == "Yes" {

			invite, ok := ctx.Value("invite").(models.Invite)
			if !ok {
				msg.Text = "There was an error getting context values. Contact the developer."
				t.bot.Send(msg)
				return errContextDataMissing
			}

			t.data.AddUserToAlert(invite.Alert, update.User)
			msg.Text = "You successfully joined the %s chanel. You will be notified on the next alert."
			ctx.Set("invite", nil)

		} else if update.Text == "No" {
			msg.Text = "You did not join."
			ctx.Set("invite", nil)
		} else {
			msg.Text = "Please answer with Yes/No"
			t.bot.Send(msg)
			return errUnexpectedUserInput
		}

		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		t.bot.Send(msg)

		return err
	})

}
