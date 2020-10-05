package telegram

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
}

// NewTelegram creates a new instance of a Telegram
func NewTelegram(config *Configuration, repository repository.Repository, webserver web.Webserver) (*Telegram, error) {

	// connect to telegram
	bot, err := tgbotapi.NewBotAPI(config.APIKey)
	if err != nil {
		return nil, err
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	tg := &Telegram{
		bot:        bot,
		repository: repository,
		config:     config,
		webserver:  webserver,
	}

	// create new dialog
	root := dialog.NewRoot()
	root.Branch(
		tg.command("start").Append(tg.newStartDialog()),
		tg.command("create").Append(tg.newCreateAlertDialog()),
		tg.command("delete").Append(tg.newDeleteDialog()),
		tg.command("info").Append(tg.newInfoDialog()),
		tg.command("alert_info").Append(tg.newAlertInfoDialog()),
		tg.command("unsubscribe").Append(tg.newAlertUnsubscribeDialog()),
		tg.command("change_alert_token").Append(tg.newAlertChangeTokenDialog()),
		tg.command("invite").Append(tg.newAlertInviteDialog()),
		tg.command("delete_invite").Append(tg.newInviteDeleteDiaolg()),
	)

	tg.conversation = dialog.NewManager(root)

	return tg, nil
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

	log.Print("Listening to telegram updates...")

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		t.processUpdate(update)
	}

	log.Print("Telegram bot shutdown")
	return nil
}

func (t *Telegram) processUpdate(update tgbotapi.Update) {

	if update.Message != nil {

		// create user if it does not exist
		user, err := t.repository.GetUserTelegram(update.Message.Chat.ID)
		if err == sql.ErrNoRows {
			user, err = t.repository.CreateUser(*models.NewUser(
				update.Message.Chat.ID,
				update.Message.Chat.UserName,
			))

			if err != nil {
				log.Printf("Error while writing userdata: %s", err.Error())
			}
		} else if err != nil {
			log.Printf("Error while reading userdata: %s", err.Error())
		}

		dialogUpdate := Update{
			Update: update,
			User:   user,
			Text:   update.Message.Text,
			ChatID: update.Message.Chat.ID,
		}

		log.Printf("Message from User:%x -> %s", dialogUpdate.User.ID, dialogUpdate.Text)

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
			t.bot.Send(msg)
			break
		case "":
			break
		default:
			err := t.conversation.Next(dialogUpdate, update.Message.Chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(dialogUpdate.ChatID, "")

				switch err {
				case dialog.ErrNoMatch:
					msg.Text = "I dont know what to do with this. Please use the provided commands or use /exit if something is not working."
				default:
					msg.Text = "We are sorry. A fatal error occured. A developer will have a look into this.\n" +
						"Unfortunately your current action was reset by this. Please try again."

					log.Printf("Telegram user(%s) error: %s", user.Username, err.Error())
				}

				t.bot.Send(msg)
			}
		}
	}
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
	message = fmt.Sprintf("\a %s: %s", alert.Name, message)

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
			err = t.sendImage(alert, fileReader)
			if err != nil {
				return err
			}
		} else {
			err = t.sendDocument(alert, fileReader)
			if err != nil {
				return err
			}
		}
	}

	// send text messages
	if len(message) > 0 {
		err := t.sendToAll(alert, message)
		if err != nil {
			return err
		}
	}

	return nil
}
