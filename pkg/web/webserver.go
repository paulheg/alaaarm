package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gofiber/fiber"
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
	endpoints map[string]endpoints.Endpoint
}

// NewWebserver creates a new Webserver
func NewWebserver(config *Configuration) Webserver {

	webApp := fiber.New(&fiber.Settings{
		DisableStartupMessage: true,
		CaseSensitive:         true,
	})

	return &FiberWebserver{
		endpoints: make(map[string]endpoints.Endpoint),
		server:    webApp,
		config:    config,
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

	alert.Get("/trigger", func(c *fiber.Ctx) {
		var err error

		token := c.Params("token")
		message := c.Query("m")

		if e, ok := w.endpoints["telegram"]; ok {
			err = e.NotifyAll(token, message, nil)
		} else {
			err = errEndpointNotFound
		}

		if err != nil {
			log.Println(err.Error())
			c.Status(http.StatusInternalServerError).SendString(err.Error())
			return
		}

		c.SendStatus(http.StatusOK)
	})

	alert.Post("/trigger", func(c *fiber.Ctx) {
		token := c.Params("token")
		message := c.Query("m")

		file, err := c.FormFile("file")
		if err != nil {
			log.Println(err.Error())
			c.Status(http.StatusInternalServerError).SendString(err.Error())
			return
		}

		if e, ok := w.endpoints["telegram"]; ok {
			err = e.NotifyAll(token, message, file)
		} else {
			err = errEndpointNotFound
		}

		if err != nil {
			log.Println(err.Error())
			c.Status(http.StatusInternalServerError).SendString(err.Error())
			return
		}

		c.SendStatus(http.StatusOK)
	})

	return nil
}

// Run runs the webserver
func (w *FiberWebserver) Run(wg *sync.WaitGroup) error {
	defer wg.Done()
	return w.server.Listen(8080)
}

// Quit shuts down the webserver
func (w *FiberWebserver) Quit() error {
	return w.server.Shutdown()
}

// AlertTriggerURL creates an URL to trigger the given alert
func (w *FiberWebserver) AlertTriggerURL(alert models.Alert, message string) string {
	message = url.QueryEscape(message)

	return fmt.Sprintf("https://%s/api/v1/alert/%s/trigger?m=%s", w.config.Domain, alert.Token, message)
}
