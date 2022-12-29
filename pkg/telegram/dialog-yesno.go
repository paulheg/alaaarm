package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

func (t *Telegram) newSelectDialog(optionA, optionB, questionKey string, onA, onB Failable) *dialog.Dialog {

	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(optionA),
			tgbotapi.NewKeyboardButton(optionB),
		),
	)

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		err := t.sendMessageWithReplyMarkup(u, questionKey, replyKeyboard)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		switch u.Text {
		case optionA:
			err := t.sendCloseKeyboardMessage(u, "answer", optionA)
			if err != nil {
				return dialog.Reset, err
			}

			return onA(u, ctx)
		case optionB:
			err := t.sendCloseKeyboardMessage(u, "answer", optionB)
			if err != nil {
				return dialog.Reset, err
			}

			return onB(u, ctx)
		default:
			err := t.sendMessage(u, "answer_with", optionA, optionB)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

	}))

}

func (t *Telegram) newYesNoDialog(onYes Failable, onNo Failable) *dialog.Dialog {

	const (
		// :check_mark:
		yesString string = "\u2714\ufe0f"

		// :cross_mark:
		noString string = "\u274c"
	)

	return t.newSelectDialog(yesString, noString, "yes_no_question", onYes, onNo)
}
