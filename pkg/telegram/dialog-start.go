package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

// newStartDialog
// dialog when user first open the bot
// invite id is sent as a parameter
func (t *Telegram) newStartDialog() *dialog.Dialog {
	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error
			update := i.(tgbotapi.Update)

			if update.Message != nil {
				var response string
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				if update.Message.IsCommand() && update.Message.Command() == "start" {
					invitationKey := update.Message.CommandArguments()

					if len(invitationKey) > 0 {
						// check if arugment passed to start argument is an invitation key
						exists, invite, err := t.data.GetInvitation(invitationKey)
						if exists && err == nil {

							response = fmt.Sprintf("Do you want to join the %s alert from %s ?",
								invite.Alert.Name,
								invite.Alert.Owner.Username,
							)
							msg.ReplyMarkup = yesNoMenuKeyboard

							// safe to context
							ctx.Set("invite", invite)
						} else {
							response = "The invitation does no longer exist."
						}
					} else {
						// normal start command
						response = "Welcome to the Alaaarm bot.\nWith this bot you can create and receive alerts."
						return dialog.ErrReset
					}

				} else {
					return dialog.ErrNoMatch
				}

				if len(response) > 0 {
					msg.Text = response
					t.bot.Send(msg)
				}
			} else {
				err = dialog.ErrNoMatch
			}

			return err
		},
	).Chain(func(i interface{}, ctx dialog.ValueStore) error {
		var err error
		update := i.(tgbotapi.Update)

		if update.Message != nil {

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if update.Message.Text == "Yes" {
				msg.Text = "You successfully joined the %s chanel. You will be notified on the next alert."

			} else if update.Message.Text == "No" {
				msg.Text = "You did not join."
			} else {
				msg.Text = "Please answer with Yes/No"
				return errUnexpectedUserInput
			}
			ctx.Set("invite", nil)

			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			t.bot.Send(msg)
		}

		return err
	})

}
