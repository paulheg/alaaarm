package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kyokomi/emoji"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/messages"
	"github.com/samber/lo"
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

// this only adds the command to be displayed as a telegram suggestion
func (t *Telegram) addCommandDefinition(command, description string, scope Scope) {

	botCommand := tgbotapi.BotCommand{
		Command:     command,
		Description: description,
	}

	if scope.IsAdmin() {
		t.commands[ADMIN_SCOPE] = append(t.commands[ADMIN_SCOPE], botCommand)
	}

	if scope.IsGroup() {
		t.commands[GROUP_SCOPE] = append(t.commands[GROUP_SCOPE], botCommand)
	}

	if scope.IsPrivate() {
		t.commands[PRIVATE_SCOPE] = append(t.commands[PRIVATE_SCOPE], botCommand)
	}
}

// Register a new command (without the slash) which marks the beginning of the typical interaction.
// It also will be registered in Telegram with a suggestion
func (t *Telegram) command(command, description string, scope Scope) *dialog.Dialog {

	t.addCommandDefinition(command, description, scope)

	// dialog that checks if the command and scope matches
	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		if u.Update.Message.Command() != command {
			return dialog.NoMatch, nil
		}

		scopeMatches := func() (bool, error) {
			// check command scope
			chat := u.Update.FromChat()

			if chat.IsGroup() || chat.IsSuperGroup() {
				if scope.IsGroup() {
					return true, nil
				}

				admin, err := t.isGroupAdmin(u)
				if err != nil {
					return false, err
				}

				if admin && scope.IsAdmin() {
					return true, nil
				}

				if err != nil {
					return true, nil
				}

			} else if u.Update.FromChat().IsPrivate() && scope.IsPrivate() {
				return true, nil
			}

			return false, nil
		}

		matches, err := scopeMatches()
		if err != nil {
			return dialog.Reset, err
		}

		if matches {
			return dialog.Next, nil
		} else {
			return dialog.NoMatch, nil
		}
	}))
}

func (t *Telegram) isGroupAdmin(update Update) (bool, error) {

	chatID := update.ChatID
	senderID := update.Update.SentFrom().ID

	admins, err := t.bot.GetChatAdministrators(tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	})

	if err != nil {
		return false, err
	}

	isAdmin := lo.ContainsBy(admins, func(user tgbotapi.ChatMember) bool {
		return user.User.ID == senderID
	})

	return isAdmin, nil
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

func (t *Telegram) emojify(s string, a ...interface{}) string {
	return fmt.Sprintf(emoji.Sprint(s), a...)
}

func (t *Telegram) escapedMessage(mode string, chatID int64, s string, a ...interface{}) tgbotapi.MessageConfig {
	escapedArguments := make([]interface{}, len(a))

	for i, unescapedArgument := range a {
		escapedArguments[i] = tgbotapi.EscapeText(mode, fmt.Sprint(unescapedArgument))
	}

	text := t.emojify(s, a...)

	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = mode
	return message
}

func (t *Telegram) sendMessageWithReplyMarkup(u Update, key string, replyMarkup interface{}, a ...interface{}) error {
	text, err := t.lookupText(u, key)
	if err != nil {
		return err
	}

	msg := t.escapedMessage(tgbotapi.ModeHTML, u.ChatID, text, a...)
	msg.ReplyMarkup = replyMarkup

	_, err = t.bot.Send(msg)

	return err
}

func (t *Telegram) escapedHTMLMessage(chatID int64, s string, a ...interface{}) tgbotapi.MessageConfig {
	return t.escapedMessage(tgbotapi.ModeHTML, chatID, s, a...)
}

func (t *Telegram) sendMessage(u Update, key string, a ...interface{}) error {
	return t.sendMessageWithReplyMarkup(u, key, nil, a...)
}

func (t *Telegram) sendCloseKeyboardMessage(u Update, key string, a ...interface{}) error {
	replayMarkup := tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}

	return t.sendMessageWithReplyMarkup(u, key, replayMarkup, a...)
}
