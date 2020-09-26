package telegram

import (
	"strings"

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

func (t *Telegram) command(s string) *dialog.Dialog {
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		if u.Update.Message.Command() != s {
			return dialog.NoMatch, nil
		}

		return dialog.Next, nil
	}))
}
