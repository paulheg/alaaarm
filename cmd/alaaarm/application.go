package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/paulheg/alaaarm/pkg/repository"
	"github.com/paulheg/alaaarm/pkg/repository/postgres"
	"github.com/paulheg/alaaarm/pkg/telegram"
	"github.com/paulheg/alaaarm/pkg/web"
)

var (
	// ErrConfigurationMissing is returned when the configuration was not read
	ErrConfigurationMissing = errors.New("The configuration needs to be loaded first")
)

// Application is the base struct
type Application struct {
	config     *Configuration
	repository repository.Repository
	telegram   *telegram.Telegram
	web        web.Webserver
}

// Configuration holds the application configurations
type Configuration struct {
	Logging    string
	Web        *web.Configuration
	Telegram   *telegram.Configuration
	Repository *repository.Configuration
}

var defaultConfig = &Configuration{
	Logging: "warn",
	Repository: &repository.Configuration{
		ConnectionString:   "./database.db",
		MigrationDirectory: "./migration",
		Database:           "sqlite",
	},
	Telegram: &telegram.Configuration{
		APIKey: "TelegramApiKey",
	},
	Web: &web.Configuration{
		Domain: "example.com",
		Port:   "3000",
	},
}

func newApplication() *Application {

	a := &Application{}

	return a
}

// LoadConfiguration loads the configuration file
func (a *Application) LoadConfiguration(path string) error {

	log.Info("Reading configuration")

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
	log.Info("Configuration was succesfully loaded")
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

	log.Infof("The default configuration file was written to %s", path)

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

	var logLevel log.Level

	switch a.config.Logging {
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	default:
		logLevel = log.WarnLevel
	}

	log.SetLevel(logLevel)

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
	a.repository = postgres.New()

	err := a.repository.InitDatabase(a.config.Repository)
	if err != nil {
		return err
	}

	return a.repository.MigrateDatabase()
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

	a.telegram, err = telegram.NewTelegram(a.config.Telegram, a.repository, a.web)

	return err
}
