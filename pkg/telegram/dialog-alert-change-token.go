package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertChangeTokenDialog() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			updatedAlert, err := t.repository.UpdateAlertToken(alert)
			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(updatedAlert, "Hello World")

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.Text = emoji.Sprintf(`The token of :bell: *%s* was changed.

Your new trigger URL is:
[%s](%s)`,
				updatedAlert.Name,
				triggerURL,
				triggerURL,
			)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))

}
