package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/paulheg/alaaarm/data"

	"github.com/labstack/echo"
	"github.com/paulheg/alaaarm/endpoints"

	_ "github.com/mattn/go-sqlite3"
	//rpio "github.com/stianeikeland/go-rpio"
)

// var (
// 	inputPin = rpio.Pin(24)
// )

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

	// database setup
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	// userdata
	userdata := data.NewTelegramUserData(db)

	// telegram setup
	telegramAPIKey := os.Getenv(apiKeyEnvKey)
	if len(telegramAPIKey) <= 0 {
		log.Fatal("Telegram API-Key is not present, set the Key to the environment variable ", apiKeyEnvKey)
	}

	telegram := endpoints.NewTelegramEndpoint(telegramAPIKey, userdata)
	go telegram.Run()

	// webservice setup
	t := &Template{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.GET("/", webHandler)
	e.Static("/static", "web/static")

	e.POST("/sendtoall", func(c echo.Context) error {
		message := c.FormValue("message")
		if len(message) > 0 {
			err := telegram.NotifyAll(message)
			if err != nil {
				return c.JSON(http.StatusOK, errorResponseMessage(err.Error()))
			}
			return c.JSON(http.StatusOK, successResponseMessage("Messages were sent."))
		}

		return c.JSON(http.StatusOK, errorResponseMessage("Message was empty."))
	})
	e.POST("/delete", func(c echo.Context) error {
		//id := c.FormValue("userid")

		userdata.DeleteUser(1)

		return nil
	})
	e.POST("/authorize", func(c echo.Context) error {
		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))

	// listen to event
	//telegram.NotifyAll("Alaaaarm")

	// err := rpio.Open()
	// if err != nil {
	// 	log.panic("unable to open gpio", err.Error())
	// }

	// defer rpio.Close()
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func webHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "main", nil)
}

func errorResponseMessage(message string) responseMessage {
	return responseMessage{
		Type:    "error",
		Message: message,
	}
}

func successResponseMessage(message string) responseMessage {
	return responseMessage{
		Type:    "success",
		Message: message,
	}
}

type responseMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
