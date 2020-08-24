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

func (t *Telegram) newSelectAlertDialog(command string, onSelection func(update Update, alert models.Alert) (string, error)) *dialog.Dialog {

	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {

			update := i.(Update)

			if update.Update.Message.Command() != command {
				return dialog.ErrNoMatch
			}

			alerts, err := t.data.GetUserTelegramAlerts(update.ChatID)
			if err != nil {
				return err
			}

			msg := tgbotapi.NewMessage(update.ChatID, "")

			if len(alerts) > 0 {
				// build keyboard
				buttons := make([]tgbotapi.KeyboardButton, 0)

				for _, alert := range alerts {
					buttons = append(buttons, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)))...)
				}

				msg.Text = "Select the alert"
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)
				t.bot.Send(msg)
			} else {

				msg.Text = "You dont have any alerts jet."
				t.bot.Send(msg)
				return dialog.ErrReset
			}

			return nil
		}).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {

			update := i.(Update)

			alerts, err := t.data.GetUserTelegramAlerts(update.ChatID)

			if err != nil {
				return err
			}

			var alert models.Alert
			foundAlert := false

			alertIdentifier := update.Text
			for _, alert = range alerts {
				if userFriendlyAlertIdentifier(alert) == alertIdentifier {
					foundAlert = true
					break
				}
			}

			if foundAlert {

				response, err := onSelection(update, alert)
				if err != nil {
					return err
				}

				msg := tgbotapi.NewMessage(update.ChatID, response)

				// reset keyboard
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
					HideKeyboard: true,
				}
				msg.ParseMode = tgbotapi.ModeMarkdown
				t.bot.Send(msg)
			} else {
				// return error, to select again
				msg := tgbotapi.NewMessage(update.ChatID, "Could not find the alert you selected")
				t.bot.Send(msg)
				return err
			}

			return nil
		})
}
