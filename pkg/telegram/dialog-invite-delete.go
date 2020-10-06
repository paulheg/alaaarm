package telegram

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newInviteDeleteDiaolg() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")

			invite, err := t.repository.GetInviteByAlertID(alert.ID)
			if err == sql.ErrNoRows {
				msg.Text = fmt.Sprintf("There was no invite to delete for the %s alert.", alert.Name)
				t.bot.Send(msg)
				return dialog.Reset, nil
			} else if err != nil {
				return dialog.Reset, err
			}

			ctx.Set("invite", invite)

			msg.Text = fmt.Sprintf(
				"Do you want to delete the invite for the %s alert?",
				alert.Name,
			)

			t.bot.Send(msg)
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

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprint(":check_mark_button: The invite was successfuly deleted.")
			t.bot.Send(msg)

			return dialog.Success, nil
		},
		// on no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprint(":cross_mark: The invite was not deleted.")
			t.bot.Send(msg)
			return dialog.Success, nil
		},
	))

}
