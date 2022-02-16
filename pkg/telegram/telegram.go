package telegram

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"mime/multipart"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/messages"
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/paulheg/alaaarm/pkg/repository"
	"github.com/paulheg/alaaarm/pkg/web"
)

const (
	ALERT_CONTEXT_KEY  = "alert"
	ALERTS_CONTEXT_KEY = "alerts"
	INVITE_CONTEXT_KEY = "invite"
)

var (
	errContextDataMissing = errors.New("context data missing")
	errDocumentIDMissing  = errors.New("telegram document fileID missing in response")
	errPhotoIDMissing     = errors.New("telegram photo fileID missing")
)

// Failable is a function used for dialog.Failable functions
type Failable func(u Update, ctx dialog.ValueStore) (dialog.Status, error)

func failable(f Failable) dialog.Failable {
	return func(i interface{}, ctx dialog.ValueStore) (dialog.Status, error) {
		update := i.(Update)
		return f(update, ctx)
	}
}

// Telegram represents the telegram interface
type Telegram struct {
	config       *Configuration
	bot          *tgbotapi.BotAPI
	repository   repository.Repository
	webserver    web.Webserver
	conversation *dialog.Manager
	log          *logrus.Entry
	commands     []tgbotapi.BotCommand
	library      messages.Library
}

// NewTelegram creates a new instance of a Telegram
func NewTelegram(config *Configuration, repository repository.Repository, webserver web.Webserver, logger *logrus.Logger, lib messages.Library) (*Telegram, error) {

	// connect to telegram
	bot, err := tgbotapi.NewBotAPI(config.APIKey)
	if err != nil {
		return nil, err
	}

	t := &Telegram{
		bot:        bot,
		repository: repository,
		config:     config,
		webserver:  webserver,
		log:        logger.WithField("service", "telegram"),
		library:    lib,
	}

	tgbotapi.SetLogger(t.log)
	bot.Debug = false
	t.log.WithField("bot_name", bot.Self.UserName).Debug("Authorized Telegram bot.")

	// create new dialog
	root := dialog.NewRoot()
	root.Branch(
		t.command("start", "Start talking to the bot").Append(t.newStartDialog()),
		t.command("create", "Create new Alert").Append(t.newCreateAlertDialog()),
		t.command("delete", "Delete a previously created invite").Append(t.newDeleteDialog()),
		t.command("info", "Get info about your created or subscribed alerts").Append(t.newInfoDialog()),
		t.command("alert_info", "Get info about an alert").Append(t.newAlertInfoDialog()),
		t.command("unsubscribe", "Unsubscribe from an alert you were invited to").Append(t.newAlertUnsubscribeDialog()),
		t.command("change_alert_token", "Change the alert URL token").Append(t.newAlertChangeTokenDialog()),
		t.command("invite", "Create an invitation link for your alert").Append(t.newAlertInviteDialog()),
		t.command("delete_invite", "Delete a previously created invite").Append(t.newInviteDeleteDiaolg()),
	)

	t.log.Debug("Configure bot commands per request")
	_, err = t.bot.Request(tgbotapi.NewSetMyCommands(t.commands...))
	if err != nil {
		t.log.WithError(err).Debug("Bot commands could not be configured per request")
		return t, err
	}

	t.conversation = dialog.NewManager(root)

	return t, nil
}

// Quit shuts down the telegram bot
func (t *Telegram) Quit() error {
	t.bot.StopReceivingUpdates()

	return nil
}

// Run runs the telegram bot
func (t *Telegram) Run(wg *sync.WaitGroup) error {
	defer wg.Done()

	updates := t.bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: 60,
		Offset:  0,
	})

	t.log.Info("Listening to telegram updates...")

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for telegramUpdate := range updates {
		err := t.processUpdate(telegramUpdate)
		if err != nil {
			t.log.WithError(err).Error("Telegram runtime error")
			t.sendErrorMessage(telegramUpdate.Message.Chat.ID)
			break
		}
	}

	t.log.Info("Telegram bot shutdown")
	return nil
}

func (t *Telegram) sendErrorMessage(chatID int64) {
	msg := t.escapedHTMLMessage(chatID, `We are sorry :smiling_face_with_tear:. 
A fatal error occured while processing your current action.
A developer will have a look into this.

Unfortunately your current action was reset by this. Please try again.`)

	t.bot.Send(msg)
}

func (t *Telegram) makeUpdate(update tgbotapi.Update) (Update, error) {

	if update.Message == nil {
		return Update{}, nil
	}

	// create user if it does not exist
	user, err := t.repository.GetUserTelegram(update.Message.Chat.ID)
	if err == sql.ErrNoRows {
		user, err = t.repository.CreateUser(*models.NewUser(
			update.Message.Chat.ID,
			update.Message.Chat.UserName,
		))

		if err != nil {
			return Update{}, fmt.Errorf("error while writing userdata: %s", err.Error())
		}
	} else if err != nil {
		return Update{}, fmt.Errorf("error while reading userdata: %s", err.Error())
	}

	dialogUpdate := Update{
		Update:     update,
		User:       user,
		Text:       update.Message.Text,
		ChatID:     update.Message.Chat.ID,
		Language:   update.Message.From.LanguageCode,
		Dictionary: t.library.Get(update.Message.From.LanguageCode),
	}

	if dialogUpdate.Dictionary.Key() != dialogUpdate.Language {
		t.log.WithField("lang", dialogUpdate.Language).Info("There is no translation for this language")
	}

	return dialogUpdate, nil
}

func (t *Telegram) processUpdate(telegramUpdate tgbotapi.Update) error {

	update, err := t.makeUpdate(telegramUpdate)
	if err != nil {
		return err
	}

	t.log.WithFields(logrus.Fields{
		"userID":  update.User.ID,
		"message": update.Text,
	}).Debug("New Message")

	// commands
	switch update.Text {
	case "/exit":
		// reset the dialog tree
		t.conversation.Reset(update.ChatID)

		err := t.sendCloseKeyboardMessage(update, "exit_message")
		if err != nil {
			return err
		}

		break
	case "":
		break
	default:
		err := t.conversation.Next(update, update.ChatID)
		if err == dialog.ErrNoMatch {

			err = t.sendMessage(update, "unknown_input")
			if err != nil {
				return err
			}

		} else if err != nil {
			return err
		}
	}

	return nil
}

func (t *Telegram) newFileReader(file multipart.FileHeader) (tgbotapi.FileReader, error) {

	f, err := file.Open()
	if err != nil {
		return tgbotapi.FileReader{}, err
	}

	return tgbotapi.FileReader{
		Name:   file.Filename,
		Reader: f,
	}, nil
}

// NotifyAll notifies all clients associated to the alert
func (t *Telegram) NotifyAll(token, message string, file *multipart.FileHeader) error {
	alert, err := t.repository.GetAlertByToken(token)

	if err != nil {
		return err
	}

	// append alert name
	message = t.emojify(":bell: %s: %s", alert.Name, message)

	// send files (documents / pictures)
	if file != nil {

		fileReader, err := t.newFileReader(*file)
		if err != nil {
			return err
		}

		isImage := false
		mimeType := file.Header.Get("Content-Type")

		switch mimeType {
		// gifs have to be send as document
		case "image/gif":
			isImage = false
			break
		default:
			if strings.HasPrefix(mimeType, "image") {
				isImage = true
			}
			break
		}

		if isImage {
			err = t.sendImageToAll(alert, fileReader, message)
			if err != nil {
				return err
			}
		} else {
			err = t.sendDocumentToAll(alert, fileReader, message)
			if err != nil {
				return err
			}
		}
	} else if len(message) > 0 {
		err := t.sendMessageToAll(alert, message)
		if err != nil {
			return err
		}
	}

	return nil
}
