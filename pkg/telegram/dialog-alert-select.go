package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func userFriendlyAlertIdentifier(alert models.Alert) string {
	return fmt.Sprintf("%d %s", alert.ID, alert.Name)
}

func (t *Telegram) newSelectSubscribedAlertDialog() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, err := t.repository.GetUserSubscribedAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		msg := tgbotapi.NewMessage(u.ChatID, "")

		if len(alerts) == 0 {
			msg.Text = "You are not subscribed to anly alerts yet."
			t.bot.Send(msg)
			return dialog.Reset, nil
		}

		ctx.Set("alerts", alerts)

		// build keyboard
		buttons := make([]tgbotapi.KeyboardButton, 0)

		for _, alert := range alerts {
			buttons = append(buttons,
				tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)),
			)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)

		msg.Text = "Select the alert"
		t.bot.Send(msg)

		return dialog.Success, nil
	})).Append(t.alertDetermination())
}

func (t *Telegram) newSelectAlertDialog() *dialog.Dialog {

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, err := t.repository.GetUserAlerts(u.User.ID)
		if err != nil {
			return dialog.Reset, err
		}

		msg := tgbotapi.NewMessage(u.ChatID, "")

		if len(alerts) == 0 {
			msg.Text = "You dont have any alerts yet."
			t.bot.Send(msg)
			return dialog.Reset, nil
		}

		// save the alerts
		ctx.Set("alerts", alerts)

		// build keyboard
		buttons := make([]tgbotapi.KeyboardButton, 0)

		for _, alert := range alerts {
			buttons = append(buttons, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)))...)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)

		msg.Text = "Select the alert"
		t.bot.Send(msg)

		return dialog.Success, nil
	})).Append(t.alertDetermination())
}

func (t *Telegram) alertDetermination() *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		alerts, ok := ctx.Value("alerts").([]models.Alert)
		if !ok {
			return dialog.Reset, errContextDataMissing
		}

		var alert models.Alert
		foundAlert := false

		alertIdentifier := u.Text
		for _, alert = range alerts {
			if userFriendlyAlertIdentifier(alert) == alertIdentifier {
				foundAlert = true
				break
			}
		}

		if !foundAlert {
			msg := tgbotapi.NewMessage(u.ChatID, "Could not find the alert you selected")
			t.bot.Send(msg)
			return dialog.Retry, nil
		}

		alert, err := t.repository.GetAlert(alert.ID)
		if err != nil {
			return dialog.Reset, err
		}

		ctx.Set("alert", alert)

		msg := tgbotapi.NewMessage(u.ChatID, "")
		msg.Text = fmt.Sprintf("You selected the '%s' alert.", alert.Name)

		// reset keyboard
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
			HideKeyboard: true,
		}
		msg.ParseMode = tgbotapi.ModeMarkdown
		t.bot.Send(msg)
		return dialog.Next, nil
	}))
}
