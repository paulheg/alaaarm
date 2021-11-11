package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertUnsubscribeDialog() *dialog.Dialog {

	return t.newSelectSubscribedAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			if u.User.ID == alert.OwnerID {
				msg := tgbotapi.NewMessage(u.ChatID, "You cant unsubscribe from your own alert. Use /delete to remove it.")
				_, err := t.bot.Send(msg)
				if err != nil {
					return dialog.Reset, err
				}

				return dialog.Success, nil
			}

			err := t.repository.RemoveUserFromAlert(alert, u.User)

			if err != nil {
				return dialog.Reset, err
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprintf(":check_mark_button: You are succesfully unsubscribed from the %s alert.", alert.Name)
			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))

}
