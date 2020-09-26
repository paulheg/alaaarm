package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertInfoDialog() *dialog.Dialog {

	return t.command("alertinfo").Append(t.newSelectAlertDialog()).
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = fmt.Sprintf(
				"Name: %s\nDescription: %s\nTrigger URL: %s\n",
				alert.Name,
				alert.Description,
				t.webserver.AlertTriggerURL(alert, "Hello World"),
			)

			t.bot.Send(msg)

			return dialog.Success, nil
		}))

}
