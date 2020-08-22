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

func (t *Telegram) newSelectDialog(command string, onSelection func(alert models.Alert) (string, error)) *dialog.Dialog {

	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {

			update := i.(tgbotapi.Update)

			if !update.Message.IsCommand() || update.Message.Command() != command {
				return dialog.ErrNoMatch
			}

			alerts, err := t.data.GetUserTelegramAlerts(update.Message.Chat.ID)

			if err != nil {
				return err
			}

			if len(alerts) > 0 {
				// build keyboard
				buttons := make([]tgbotapi.KeyboardButton, 0)

				for _, alert := range alerts {
					buttons = append(buttons, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(userFriendlyAlertIdentifier(alert)))...)
				}

				keyboard := tgbotapi.NewReplyKeyboard(buttons)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select the alert")
				msg.ReplyMarkup = keyboard

				t.bot.Send(msg)
			} else {

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You dont have any alerts jet.")
				t.bot.Send(msg)
				return dialog.ErrReset
			}

			return nil
		}).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {

			update := i.(tgbotapi.Update)

			alerts, err := t.data.GetUserTelegramAlerts(update.Message.Chat.ID)

			if err != nil {
				return err
			}

			var alert models.Alert
			foundAlert := false

			alertIdentifier := update.Message.Text
			for _, alert = range alerts {
				if userFriendlyAlertIdentifier(alert) == alertIdentifier {
					foundAlert = true
					break
				}
			}

			if foundAlert {

				response, err := onSelection(alert)
				if err != nil {
					return err
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
					HideKeyboard: true,
				}
				msg.ParseMode = tgbotapi.ModeMarkdown
				t.bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Could not find the alert you selected")
				t.bot.Send(msg)
				return err
			}

			// get alert object and write to value store
			// reset keyboard

			// return error, to select again

			return nil
		})
}
