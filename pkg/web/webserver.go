package web

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"net/http"
	"net/url"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/paulheg/alaaarm/pkg/endpoints"
	"github.com/paulheg/alaaarm/pkg/models"
)

var (
	errEndpointNotFound = errors.New("The endpoint was not found")
)

// Webserver interface
type Webserver interface {
	InitializeWebserver() error
	Run(wg *sync.WaitGroup) error
	Quit() error
	AlertTriggerURL(alert models.Alert, message string) string
	RegisterEndpoint(name string, endpoint endpoints.Endpoint)
}

// FiberWebserver represents the fiber webinterface for this application
type FiberWebserver struct {
	config    *Configuration
	server    *fiber.App
	log       *logrus.Entry
	endpoints map[string]endpoints.Endpoint
}

// NewWebserver creates a new Webserver
func NewWebserver(config *Configuration, log *logrus.Logger) Webserver {

	webApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		CaseSensitive:         true,
	})

	return &FiberWebserver{
		endpoints: make(map[string]endpoints.Endpoint),
		server:    webApp,
		config:    config,
		log:       log.WithField("service", "webserver"),
	}

}

// RegisterEndpoint registers an endpoint to the webserver
func (w *FiberWebserver) RegisterEndpoint(name string, endpoint endpoints.Endpoint) {
	w.endpoints[name] = endpoint
}

// InitializeWebserver initializes the webserver
func (w *FiberWebserver) InitializeWebserver() error {

	// REST API
	api := w.server.Group("/api")
	v1 := api.Group("/v1")

	alert := v1.Group("/alert/:token")

	alert.Get("/trigger", func(c *fiber.Ctx) error {
		var err error

		token := c.Params("token")
		message := c.Query("m")

		webLogger := w.log.WithFields(logrus.Fields{
			"alert_token": token,
			"message":     message,
			"ip":          c.IP(),
			"user_agent":  string(c.Request().Header.UserAgent()),
		})

		webLogger.Debug("Triggering message over web interface")

		if e, ok := w.endpoints["telegram"]; ok {
			err = e.NotifyAll(token, message, nil)
		} else {
			err = errEndpointNotFound
		}

		if err == sql.ErrNoRows {
			return c.SendStatus(http.StatusNotFound)
		} else if err != nil {
			webLogger.Error(err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.SendStatus(http.StatusOK)
	})

	alert.Post("/trigger", func(c *fiber.Ctx) error {
		token := c.Params("token")
		message := c.Query("m")

		webLogger := w.log.WithFields(logrus.Fields{
			"alert_token": token,
			"message":     message,
			"ip":          c.IP(),
			"user_agent":  c.Request().Header.UserAgent(),
		})

		webLogger.Debug("Triggering message over web interface")

		file, err := c.FormFile("file")
		if err != nil {
			webLogger.Error(err)
			return c.SendStatus(http.StatusInternalServerError)
		}

		if e, ok := w.endpoints["telegram"]; ok {
			err = e.NotifyAll(token, message, file)
		} else {
			err = errEndpointNotFound
		}

		if err == sql.ErrNoRows {
			return c.SendStatus(http.StatusNotFound)
		} else if err != nil {
			webLogger.Error(err)
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.SendStatus(http.StatusOK)
	})

	return nil
}

// Run runs the webserver
func (w *FiberWebserver) Run(wg *sync.WaitGroup) error {
	defer wg.Done()

	w.log.Info("Webserver listening...")

	return w.server.Listen(":" + w.config.Port)
}

// Quit shuts down the webserver
func (w *FiberWebserver) Quit() error {
	err := w.server.Shutdown()
	if err != nil {
		return err
	}
	w.log.Info("Webserver shutdown")

	return nil
}

// AlertTriggerURL creates an URL to trigger the given alert
func (w *FiberWebserver) AlertTriggerURL(alert models.Alert, message string) string {
	message = url.QueryEscape(message)

	return fmt.Sprintf("https://%s/api/v1/alert/%s/trigger?m=%s", w.config.Domain, alert.Token, message)
}
