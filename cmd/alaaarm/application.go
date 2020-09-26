package main

import (
	"errors"
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

var (
	// ErrConfigurationMissing is returned when the configuration was not read
	ErrConfigurationMissing = errors.New("The configuration needs to be loaded first")
)

// Application is the base struct
type Application struct {
	config   *Configuration
	data     data.Data
	telegram *telegram.Telegram
	web      web.Webserver
}

// Configuration holds the application configurations
type Configuration struct {
	Web      *web.Configuration
	Telegram *telegram.Configuration
	Data     *data.Configuration
}

func newApplication() *Application {

	a := &Application{}

	return a
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
			Domain: "alaaarm.me",
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

// Initialize initializes the application
func (a *Application) Initialize() error {
	var err error

	if a.config == nil {
		return ErrConfigurationMissing
	}

	err = a.initializeDatabase()
	if err != nil {
		return err
	}

	err = a.initializeWebserver()
	if err != nil {
		return err
	}

	err = a.initializeTelegram()
	if err != nil {
		return err
	}

	a.web.RegisterEndpoint("telegram", a.telegram)

	return nil
}

// InitializeDatabase initializes the database
func (a *Application) initializeDatabase() error {
	a.data = data.NewGormData(a.config.Data)

	err := a.data.InitDatabase()
	if err != nil {
		return err
	}

	return a.data.MigrateDatabase()
}

// initializeWebserver initializes the webserver
func (a *Application) initializeWebserver() error {
	a.web = web.NewWebserver(a.config.Web)
	err := a.web.InitializeWebserver()

	return err
}

// initializeTelegram initializes telegram
func (a *Application) initializeTelegram() error {
	var err error

	a.telegram, err = telegram.NewTelegram(a.config.Telegram, a.data, a.web)

	return err
}
