package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newYesNoDialog(onYes Failable, onNo Failable) *dialog.Dialog {

	const (
		// :check_mark:
		yesString string = "\u2714\ufe0f"

		// :cross_mark:
		noString string = "\u274c"
	)

	yesNoReplyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(yesString),
			tgbotapi.NewKeyboardButton(noString),
		),
	)

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		err := t.sendMessageWithReplyMarkup(u, "yes_no_question", yesNoReplyKeyboard)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		switch u.Text {
		case yesString:
			err := t.sendCloseKeyboardMessage(u, "answer_yes")
			if err != nil {
				return dialog.Reset, err
			}

			return onYes(u, ctx)
		case noString:
			err := t.sendCloseKeyboardMessage(u, "answer_no")
			if err != nil {
				return dialog.Reset, err
			}

			return onNo(u, ctx)
		default:
			err := t.sendMessage(u, "answer_with_yes_no", yesString, noString)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

	}))
}
