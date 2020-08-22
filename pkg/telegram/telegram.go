package telegram

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/data"
	"github.com/paulheg/alaaarm/pkg/dialog"
)

var (
	errUserAlreadyExists   = errors.New("user already exists")
	errInvalidName         = errors.New("invalid name")
	errInvalidDescription  = errors.New("invalid description")
	errContextDataMissing  = errors.New("context data missing")
	errDocumentIDMissing   = errors.New("telegram document fileID missing in response")
	errPhotoIDMissing      = errors.New("telegram photo fileID missing")
	errUnexpectedUserInput = errors.New("the userinput was not expected and could not be processed")
)

// Keyboards
var (
	yesNoMenuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Yes"),
			tgbotapi.NewKeyboardButton("No"),
		),
	)
)

// Telegram represents the telegram interface
type Telegram struct {
	config       Configuration
	bot          *tgbotapi.BotAPI
	data         data.Data
	conversation *dialog.Manager
	quit         chan bool
}

// NewTelegram creates a new instance of a Telegram
func NewTelegram(config Configuration, data data.Data) *Telegram {

	// connect to telegram
	bot, err := tgbotapi.NewBotAPI(config.APIKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	tg := &Telegram{
		bot:    bot,
		data:   data,
		quit:   make(chan bool),
		config: config,
	}

	// create new dialog
	root := dialog.Root()
	root.Branch(
		tg.newCreateAlertDialog(),
		tg.newInfoDialog(),
		tg.newStartDialog(),
		tg.newDeleteDialog(),
		tg.newAlertInviteDialog(),
	)

	tg.conversation = dialog.NewManager(root)

	return tg
}

// Quit shuts down the telegram bot
func (t *Telegram) Quit() error {
	t.bot.StopReceivingUpdates()
	t.quit <- true

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

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {

		// TODO: FIX THIS
		// if <-t.quit {
		// 	break
		// }

		t.processUpdate(update)
	}

	log.Print("Telegram bot shutdown")
	return nil
}

func (t *Telegram) processUpdate(update tgbotapi.Update) {
	var msg tgbotapi.MessageConfig

	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// commands
		switch update.Message.Text {
		case "/exit":
			// reset the dialog tree
			t.conversation.Reset(update.Message.Chat.ID)

			// reset keyboard
			msg.Text = "You are back at the start. Send me a command to continue."
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			t.bot.Send(msg)
			break
		case "":
			break
		default:
			err := t.conversation.Next(update, update.Message.Chat.ID)
			if err != nil {
				log.Printf("Telegram user(%s) error: %s", update.Message.Chat.UserName, err.Error())
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
	exists, alert, err := t.data.GetAlertWithToken(token)

	if err != nil {
		return err
	}

	if exists {
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

	}

	return nil
}
