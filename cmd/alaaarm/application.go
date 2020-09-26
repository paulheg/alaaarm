package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"sync"

	"github.com/paulheg/alaaarm/pkg/data"
	"github.com/paulheg/alaaarm/pkg/telegram"
	"github.com/paulheg/alaaarm/pkg/web"
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

var defaultConfig = &Configuration{
	Data: &data.Configuration{
		ConnectionString: "./database.db",
	},
	Telegram: &telegram.Configuration{
		APIKey: "TelegramApiKey",
	},
	Web: &web.Configuration{
		Domain: "example.com",
	},
}

func newApplication() *Application {

	a := &Application{}

	return a
}

// LoadConfiguration loads the configuration file
func (a *Application) LoadConfiguration(path string) error {

	log.Println("Reading configuration")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	config := defaultConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	a.config = config
	log.Println("Configuration was succesfully loaded")
	return nil
}

// CreateConfiguration creates the defaul configuration file
func (a *Application) CreateConfiguration(path string) error {

	fileContent, err := json.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, fileContent, 0755)
	if err != nil {
		return err
	}

	log.Printf("The default configuration file was written to %s", path)

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
