package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	//rpio "github.com/stianeikeland/go-rpio"
)

// var (
// 	inputPin = rpio.Pin(24)
// )

var (
	errUserAlreadyExists = errors.New("user already exists")
)

const (
	apiKeyEnvKey = "ALARM_TELEGRAM_BOT_API_KEY"
)

const (
	sqlCreateUserTable = `CREATE TABLE 'USERDATA' (
		'ID'	INTEGER NOT NULL UNIQUE,
		'USERNAME'	TEXT NOT NULL,
		'AUTHORIZED'	BOOLEAN NOT NULL DEFAULT 0 CHECK(AUTHORIZED in ( 0 , 1 )),
		PRIMARY KEY('ID')
	);
	`
)

func main() {

	// open database
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	// log.Println("opening gpio")
	// err := rpio.Open()
	// if err != nil {
	// 	log.panic("unable to open gpio", err.Error())
	// }

	// defer rpio.Close()

	// connect to telegram
	telegramAPIKey := os.Getenv(apiKeyEnvKey)
	if len(telegramAPIKey) <= 0 {
		log.Fatal("Telegram API-Key is not present, set the Key to the environment variable ", apiKeyEnvKey)
	}

	bot, err := tgbotapi.NewBotAPI(telegramAPIKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

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
			err := addUser(db, update.Message.Chat.ID, update.Message.Chat.UserName)
			var response string
			if err == errUserAlreadyExists {
				response = "Du existierst bereits in der Datenbank."
			} else {
				response = "Du wirst für eine Authorisierung überprüft."
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			bot.Send(msg)

		case "/unsubscribe":
			err := deleteUser(db, update.Message.Chat.ID)
			var response string
			if err != nil {
				response = "Beim versuch dich aus der Datenbank zu entfernen ist ein fehler aufgetreten. Versuche später noch einmal."
			} else {
				response = "Du wirst in Zukunft nicht mehr alarmiert."
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			bot.Send(msg)

		case "/status":
			var response string
			authorized := isAuthorized(db, update.Message.Chat.ID)
			if authorized {
				response = "Du wird beim nächsten Alarm alarmiert."
			} else {
				response = "Du wirst nicht alarmiert, entweder hast du noch nicht aboniert, oder wurdest noch nicht freigeschaltet."
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			bot.Send(msg)

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Damit kann ich dir leider nicht helfen.")
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}

	}

}

func addUser(db *sql.DB, id int64, username string) error {
	if userExists(db, id) {
		return errUserAlreadyExists
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec("insert into USERDATA(ID, USERNAME, AUTHORIZED) values(?, ?, ?)", id, username, false)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
	return nil
}

func deleteUser(db *sql.DB, id int64) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec("delete from USERDATA where USERDATA.ID = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
	return nil
}

func isAuthorized(db *sql.DB, id int64) bool {
	row := db.QueryRow("select AUTHORIZED from USERDATA where USERDATA.ID = ? limit 1", id)

	var authorized bool = false

	err := row.Scan(&authorized)

	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	return authorized
}

func userExists(db *sql.DB, ID int64) bool {
	err := db.QueryRow("select ID from USERDATA where USERDATA.ID = ? limit 1", ID).Scan(&ID)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}

		return false
	}

	return true
}

func notifyAll(db *sql.DB, bot *tgbotapi.BotAPI, msg string) {
	rows, err := db.Query("select ID from USERDATA where USERDATA.AUTHORIZED = 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var ID int64
		err := rows.Scan(&ID)
		if err != nil {
			log.Fatal(err)
		}

		msg := tgbotapi.NewMessage(ID, msg)
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

}
