package web

import (
	"log"
	"net/http"
	"sync"

	"github.com/gofiber/fiber"
	"github.com/paulheg/alaaarm/pkg/endpoints"
)

// Webserver represents the webinterface for this application
type Webserver struct {
	config   Configuration
	server   *fiber.App
	endpoint endpoints.Endpoint
}

// NewWebserver creates a new Webserver
func NewWebserver(config Configuration, endpoint endpoints.Endpoint) *Webserver {

	webApp := fiber.New(&fiber.Settings{
		DisableStartupMessage: true,
		CaseSensitive:         true,
	})

	return &Webserver{
		endpoint: endpoint,
		server:   webApp,
	}

}

// InitializeWebserver initializes the webserver
func (w *Webserver) InitializeWebserver() error {

	// REST API
	api := w.server.Group("/api")
	v1 := api.Group("/v1")

	alert := v1.Group("/alert/:token")

	alert.Get("/trigger", func(c *fiber.Ctx) {
		token := c.Params("token")
		message := c.Query("m")

		err := w.endpoint.NotifyAll(token, message, nil)

		status := http.StatusOK

		if err != nil {
			log.Println(err.Error())
			status = http.StatusInternalServerError
		}

		c.SendStatus(status)
	})

	alert.Post("/trigger", func(c *fiber.Ctx) {
		token := c.Params("token")
		message := c.Query("m")

		file, err := c.FormFile("file")

		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
			log.Println(err.Error())
		} else {
			err = w.endpoint.NotifyAll(token, message, file)
			if err != nil {
				status = http.StatusInternalServerError
				log.Println(err.Error())
			}
		}

		c.SendStatus(status)
	})

	return nil
}

// Run runs the webserver
func (w *Webserver) Run(wg *sync.WaitGroup) error {
	defer wg.Done()
	return w.server.Listen(8080)
}

// Quit shuts down the webserver
func (w *Webserver) Quit() error {
	return w.server.Shutdown()
}
