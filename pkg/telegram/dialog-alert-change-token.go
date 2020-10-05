package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newChangeDialogTokenDialog() *dialog.Dialog {

	return t.command("changeToken").Append(t.newSelectAlertDialog()).
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			updatedAlert, err := t.repository.UpdateAlertToken(alert)
			if err != nil {
				return dialog.Reset, err
			}

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = fmt.Sprintf("The token of %s was changed.\n\nYour knew URL is:\n\n%s",
				updatedAlert.Name,
				t.webserver.AlertTriggerURL(updatedAlert, "Hello World"),
			)

			return dialog.Success, nil
		}))

}
