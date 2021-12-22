package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
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

func (t *Telegram) escapedHTMLLookup(chatID int64, key string, a ...interface{}) tgbotapi.MessageConfig {
	text, err := t.dictionary.Lookup(key)
	if err != nil {
		return tgbotapi.MessageConfig{}
	}

	return t.escapedHTMLMessage(chatID, text, a)
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
