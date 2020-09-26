package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newYesNoDialog(onYes Failable, onNo Failable) *dialog.Dialog {

	yesNoReplyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Yes"),
			tgbotapi.NewKeyboardButton("No"),
		),
	)

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "Yes / No ?")
		msg.ReplyMarkup = yesNoReplyKeyboard
		t.bot.Send(msg)
		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		switch strings.ToLower(u.Text) {
		case "yes":
			msg.Text = "You answered Yes."
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			t.bot.Send(msg)

			return onYes(u, ctx)
		case "no":
			msg.Text = "You answered No."
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			t.bot.Send(msg)

			return onNo(u, ctx)
		default:
			msg.Text = "Please answer with Yes/No"
			t.bot.Send(msg)
			return dialog.Retry, nil
		}

	}))
}
