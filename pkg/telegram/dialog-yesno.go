package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
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
		_, err := t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		switch strings.ToLower(u.Text) {
		case "yes":
			msg.Text = emoji.Sprint(":check_mark: You answered Yes.")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return onYes(u, ctx)
		case "no":
			msg.Text = emoji.Sprint(":cross_mark: You answered No.")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return onNo(u, ctx)
		default:
			msg.Text = "Please answer with Yes/No"
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

	}))
}
