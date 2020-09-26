package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertUnsubscribeDialog() *dialog.Dialog {

	return t.command("unsubscribe").Append(t.newSelectSubscribedAlertDialog()).
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			if u.User.ID == alert.OwnerID {
				msg := tgbotapi.NewMessage(u.ChatID, "You cant unsubscribe from your own alert. Use /delete to remove it.")
				t.bot.Send(msg)
				return dialog.Success, nil
			}

			err := t.data.RemoveUserFromAlert(alert, u.User)

			if err != nil {
				return dialog.Reset, err
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = fmt.Sprintf("You are succesfully unsubscribed from the %s alert.", alert.Name)
			t.bot.Send(msg)
			return dialog.Success, nil
		}))

}
