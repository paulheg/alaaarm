package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
