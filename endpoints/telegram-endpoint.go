package endpoints

import (
	"errors"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/data"
)

var (
	errUserAlreadyExists = errors.New("user already exists")
)

type telegram struct {
	bot      *tgbotapi.BotAPI
	userdata data.Userdata
}

// NewTelegramEndpoint creates a new instance of a Telegram Endpoint
func NewTelegramEndpoint(apiKey string, userdata data.Userdata) EndpointService {

	// connect to telegram
	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	tg := &telegram{
		bot:      bot,
		userdata: userdata,
	}

	return tg
}

func (t *telegram) Run() error {

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
		if update.Message == nil {
			continue
		}
		// Add logic here

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		//userId := update.Message.From.ID

		// commands
		switch update.Message.Text {
		case "/subscribe":
			err := t.userdata.AddUser(update.Message.Chat.ID, update.Message.Chat.UserName)
			var response string
			if err == errUserAlreadyExists {
				response = "Du existierst bereits in der Datenbank."
			} else {
				response = "Du wirst für eine Authorisierung überprüft."
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			t.bot.Send(msg)

		case "/unsubscribe":
			err := t.userdata.DeleteUser(update.Message.Chat.ID)
			var response string
			if err != nil {
				response = "Beim versuch dich aus der Datenbank zu entfernen ist ein fehler aufgetreten. Versuche später noch einmal."
			} else {
				response = "Du wirst in Zukunft nicht mehr alarmiert."
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			t.bot.Send(msg)

		case "/status":
			var response string
			authorized, err := t.userdata.IsUserAuthorized(update.Message.Chat.ID)
			if err != nil {
				log.Print(err)
				response = "Beim abruf der Daten ist ein Fehler aufgetreten, versuche es später noch einmal."
			} else {
				if authorized {
					response = "Du wird beim nächsten Alarm alarmiert."
				} else {
					response = "Du wirst nicht alarmiert, entweder hast du noch nicht aboniert, oder wurdest noch nicht freigeschaltet."
				}
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			t.bot.Send(msg)

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Damit kann ich dir leider nicht helfen.")
			msg.ReplyToMessageID = update.Message.MessageID

			t.bot.Send(msg)
		}

	}

	return nil
}

func (t *telegram) NotifyAll(message string) error {
	users, err := t.userdata.GetAllAuthorizedUsers()

	for _, id := range users {
		msg := tgbotapi.NewMessage(id, message)
		_, err = t.bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
