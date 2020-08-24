package main

import (
	"log"
	"os"
	"sync"

	"github.com/paulheg/alaaarm/pkg/data"
	"github.com/paulheg/alaaarm/pkg/telegram"
	"github.com/paulheg/alaaarm/pkg/web"
)

// TODO: Replace with configuration file
const (
	telegramAPIKeyEnvKey = "ALARM_TELEGRAM_BOT_API_KEY"
)

// Configuration holds the application configurations
type Configuration struct {
	Web      *web.Configuration
	Telegram *telegram.Configuration
	Data     *data.Configuration
}

// Application is the base struct
type Application struct {
	config   *Configuration
	data     data.Data
	telegram *telegram.Telegram
	web      web.Webserver
}

func newApplication() *Application {

	a := &Application{}

	return a
}

// InitializeDatabase initializes the database
func (a *Application) InitializeDatabase() error {
	a.data = data.NewGormData(*a.config.Data)
	return a.data.InitDatabase()
}

// InitializeWebserver initializes the webserver
func (a *Application) InitializeWebserver() error {
	a.web = web.NewWebserver(*a.config.Web, a.telegram)
	err := a.web.InitializeWebserver()

	return err
}

// InitializeTelegram initializes telegram
func (a *Application) InitializeTelegram() error {
	a.telegram = telegram.NewTelegram(*a.config.Telegram, a.data)

	return nil
}

// LoadConfiguration loads the configuration file
func (a *Application) LoadConfiguration(path string) error {

	//TODO: read file

	// telegram setup
	telegramAPIKey := os.Getenv(telegramAPIKeyEnvKey)
	if len(telegramAPIKey) <= 0 {
		log.Fatal("Telegram API-Key is not present, set the Key to the environment variable ", telegramAPIKeyEnvKey)
	}

	a.config = &Configuration{
		Web: &web.Configuration{
			Domain: "alaaarm.com",
		},
		Telegram: &telegram.Configuration{
			APIKey: telegramAPIKey,
			Name:   "AlaarmAlaaarmBot",
		},
		Data: &data.Configuration{
			ConnectionString: "../../database.db",
		},
	}

	return nil
}

// Run starts the application
func (a *Application) Run() {
	var wg sync.WaitGroup

	// run telegram
	wg.Add(1)
	go a.telegram.Run(&wg)

	// run the webserver
	wg.Add(1)
	go a.web.Run(&wg)

	wg.Wait()
}

// Quit shuts down the application
func (a *Application) Quit() error {
	err := a.telegram.Quit()
	if err != nil {
		return err
	}

	err = a.web.Quit()
	if err != nil {
		return err
	}

	return nil
}
