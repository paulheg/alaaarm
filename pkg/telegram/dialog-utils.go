package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/messages"
	"github.com/sirupsen/logrus"
)

func (t *Telegram) on(s string, caseSensitive bool) *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		comparer := u.Text

		if !caseSensitive {
			s = strings.ToLower(s)
			comparer = strings.ToLower(comparer)
		}

		if comparer != s {
			return dialog.NoMatch, nil
		}

		return dialog.Next, nil
	}))
}

// Register a new command (without the slash) which marks the beginning of the typical interaction
func (t *Telegram) command(command, description string) *dialog.Dialog {
	t.commands = append(t.commands, tgbotapi.BotCommand{
		Command:     command,
		Description: description,
	})

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		if u.Update.Message.Command() != command {
			return dialog.NoMatch, nil
		}

		return dialog.Next, nil
	}))
}

func (t *Telegram) lookupText(u Update, key string) (string, error) {

	text, err := u.Dictionary.Lookup(key)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"lang": u.Language,
			"key":  key,
		}).Warning("there is no translation for the given key in the selected language, switching to default language")

		text, err = t.library.Default().Lookup(key)
		if err != nil {
			t.log.WithFields(logrus.Fields{
				"lang": u.Language,
				"key":  key,
			}).Error("There is no default translation for the given key")

			return "", messages.ErrMessageNotFound
		}
	}

	return text, nil
}

func (t *Telegram) sendMessageWithReplayMarkup(u Update, key string, replyMarkup interface{}, a ...interface{}) error {
	text, err := t.lookupText(u, key)
	if err != nil {
		return err
	}

	msg := t.escapedHTMLMessage(u.ChatID, text, a)
	msg.ReplyMarkup = replyMarkup

	_, err = t.bot.Send(msg)

	return err
}

func (t *Telegram) sendMessage(u Update, key string, a ...interface{}) error {
	return t.sendMessageWithReplayMarkup(u, key, nil, a)
}

func (t *Telegram) sendCloseKeyboardMessage(u Update, key string, a ...interface{}) error {
	replayMarkup := tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}

	return t.sendMessageWithReplayMarkup(u, key, replayMarkup, a)
}

func (t *Telegram) escapedHTMLMessage(chatID int64, s string, a ...interface{}) tgbotapi.MessageConfig {
	const mode = tgbotapi.ModeHTML

	escapedArguments := make([]interface{}, len(a))

	for i, unescapedArgument := range a {
		escapedArguments[i] = tgbotapi.EscapeText(mode, fmt.Sprint(unescapedArgument))
	}

	text := emoji.Sprintf(s, escapedArguments...)

	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = mode
	return message
}

func (t *Telegram) escapedMarkdownMessage(chatID int64, s string, a ...interface{}) tgbotapi.MessageConfig {
	const mode = tgbotapi.ModeMarkdown

	escapedArguments := make([]interface{}, len(a))

	for i, unescapedArgument := range a {
		escapedArguments[i] = tgbotapi.EscapeText(mode, fmt.Sprint(unescapedArgument))
	}

	text := emoji.Sprintf(s, escapedArguments...)

	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = mode
	return message
}
