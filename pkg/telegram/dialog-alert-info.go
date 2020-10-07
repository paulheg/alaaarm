package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertInfoDialog() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			receiver, err := t.repository.GetAlertReceiver(alert)
			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(alert, "Hello World")

			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.Text = emoji.Sprintf(`:bell: *Alert Info* :megaphone:

*Name:* %s
*Description:* %s
*Subscribed Users:* :moai:%v 

*Trigger URL:* 
[%s](%s)`,
				alert.Name,
				alert.Description,
				len(receiver),
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
