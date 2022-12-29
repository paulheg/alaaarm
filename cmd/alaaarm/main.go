package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	LOG_LEVEL_FLAG_NAME   = "loglevel"
	CONFIG_FILE_FLAG_NAME = "config"
)

func main() {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	log := logrus.New()
	log.SetOutput(os.Stdout)

	defaultFlags := []cli.Flag{
		&cli.PathFlag{
			Name:    CONFIG_FILE_FLAG_NAME,
			Usage:   "Load configuration from `FILE`.json",
			Value:   "./config.json",
			EnvVars: []string{"CONFIG_PATH"},
		},
		&cli.StringFlag{
			Name:    LOG_LEVEL_FLAG_NAME,
			Usage:   "Set the level of detail the logs contain (debug, info, warn, error)",
			Value:   "info",
			EnvVars: []string{"LOG_LEVEL"},
		},
	}

	app := &cli.App{
		Name:                 "Alaaarm Bot",
		Description:          "Server that is in control of the Telegram bot and the web-interface triggering alerts",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "run",
				Description: "run the server",
				Action: func(c *cli.Context) error {
					setLogLevel(log, c.String(LOG_LEVEL_FLAG_NAME))
					runCmd(log, c.Path(CONFIG_FILE_FLAG_NAME), gracefulStop)
					return nil
				},
				Flags: defaultFlags,
			},
			{
				Name: "config",
				Subcommands: []*cli.Command{
					{
						Name:        "check",
						Description: "Check the configuration file",
						Flags:       defaultFlags,
						Action: func(c *cli.Context) error {
							setLogLevel(log, c.String(LOG_LEVEL_FLAG_NAME))
							application := newApplication(log)

							err := application.LoadConfiguration(c.Path(CONFIG_FILE_FLAG_NAME))
							if err != nil {
								log.WithError(err).Fatal("An error occured while reading the configuration: ")
								return err
							}
							log.Info("Config file seems correct")
							return nil
						},
					},
					{
						Name:        "create",
						Description: "Create a default configuration file",
						Flags:       defaultFlags,
						Action: func(c *cli.Context) error {
							setLogLevel(log, c.String(LOG_LEVEL_FLAG_NAME))
							application := newApplication(log)

							log.Info("Writing default configuration")

							err := application.CreateConfiguration(c.Path(CONFIG_FILE_FLAG_NAME))
							if err != nil {
								log.WithError(err).Fatal("There was an error writing the default configuration")
							}
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func setLogLevel(log *logrus.Logger, logLevel string) {
	switch logLevel {
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

func runCmd(log *logrus.Logger, configurationPath string, quit chan os.Signal) {
	application := newApplication(log)

	err := application.LoadConfiguration(configurationPath)
	if err != nil {
		log.WithError(err).Fatal("An error occured while reading the configuration")
	}

	err = application.Initialize()
	if err != nil {
		log.WithError(err).Fatal("An error occured while initializing the application")
	}

	go func() {
		sig := <-quit
		log.WithField("signal", sig).Info("Caught signal")
		err := application.Quit()
		if err != nil {
			log.WithError(err).Fatal("A fatal error occured while trying to shut down the application")
		}

	}()

	application.Run()

	log.Info("Application finished")
}
