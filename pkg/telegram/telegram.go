package telegram

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kyokomi/emoji"
	"github.com/sirupsen/logrus"

	"mime/multipart"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/paulheg/alaaarm/pkg/repository"
	"github.com/paulheg/alaaarm/pkg/web"
)

var (
	errUserAlreadyExists   = errors.New("user already exists")
	errContextDataMissing  = errors.New("context data missing")
	errDocumentIDMissing   = errors.New("telegram document fileID missing in response")
	errPhotoIDMissing      = errors.New("telegram photo fileID missing")
	errUnexpectedUserInput = errors.New("the userinput was not expected and could not be processed")
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
}

// NewTelegram creates a new instance of a Telegram
func NewTelegram(config *Configuration, repository repository.Repository, webserver web.Webserver, logger *logrus.Logger) (*Telegram, error) {

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
	}

	tgbotapi.SetLogger(t.log)
	bot.Debug = false
	t.log.WithField("bot_name", bot.Self.UserName).Debug("Authorized Telegram bot.")

	// create new dialog
	root := dialog.NewRoot()
	root.Branch(
		t.command("start").Append(t.newStartDialog()),
		t.command("create").Append(t.newCreateAlertDialog()),
		t.command("delete").Append(t.newDeleteDialog()),
		t.command("info").Append(t.newInfoDialog()),
		t.command("alert_info").Append(t.newAlertInfoDialog()),
		t.command("unsubscribe").Append(t.newAlertUnsubscribeDialog()),
		t.command("change_alert_token").Append(t.newAlertChangeTokenDialog()),
		t.command("invite").Append(t.newAlertInviteDialog()),
		t.command("delete_invite").Append(t.newInviteDeleteDiaolg()),
	)

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	t.log.Info("Listening to telegram updates...")

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		err = t.processUpdate(update)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "We are sorry. A fatal error occured. A developer will have a look into this.\n" +
				"Unfortunately your current action was reset by this. Please try again."
			t.bot.Send(msg)

			t.log.WithError(err).Error("Telegram runtime error")
		}

	}

	t.log.Info("Telegram bot shutdown")
	return nil
}

func (t *Telegram) processUpdate(update tgbotapi.Update) error {

	if update.Message != nil {

		// create user if it does not exist
		user, err := t.repository.GetUserTelegram(update.Message.Chat.ID)
		if err == sql.ErrNoRows {
			user, err = t.repository.CreateUser(*models.NewUser(
				update.Message.Chat.ID,
				update.Message.Chat.UserName,
			))

			if err != nil {
				return fmt.Errorf("Error while writing userdata: %s", err.Error())
			}
		} else if err != nil {
			return fmt.Errorf("Error while reading userdata: %s", err.Error())

		}

		dialogUpdate := Update{
			Update: update,
			User:   user,
			Text:   update.Message.Text,
			ChatID: update.Message.Chat.ID,
		}

		t.log.WithFields(logrus.Fields{
			"userID":  dialogUpdate.User.ID,
			"message": dialogUpdate.Text,
		}).Debug("New Message")

		// commands
		switch update.Message.Text {
		case "/exit":
			// reset the dialog tree
			t.conversation.Reset(update.Message.Chat.ID)

			// reset keyboard
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"You are back at the start. Send me a command to continue.",
			)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			_, err := t.bot.Send(msg)
			if err != nil {
				return err
			}

			break
		case "":
			break
		default:
			err := t.conversation.Next(dialogUpdate, update.Message.Chat.ID)
			if err == dialog.ErrNoMatch {
				msg := tgbotapi.NewMessage(dialogUpdate.ChatID, "")
				msg.Text = "I dont know what to do with this. Please use the provided commands or use /exit if something is not working."
				_, err = t.bot.Send(msg)
				if err != nil {
					return err
				}

			} else if err != nil {
				return err
			}
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
		Size:   file.Size,
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
	message = emoji.Sprintf(":bell: %s: %s", alert.Name, message)

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
