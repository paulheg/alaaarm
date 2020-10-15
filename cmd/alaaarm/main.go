package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

var version = flag.Bool("version", false, "Show alaaarm version")
var configPath = flag.String("config", "./config/config.json", "Path to configuration file")
var logLevel = flag.String("loglevel", "info", "Loggin Level (debug, info, warn, error)")

var log = logrus.New()

func main() {

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	log.SetOutput(os.Stdout)
	setLogLevel()

	flag.Usage = func() {
		helpCmd()
	}
	flag.Parse()

	command := flag.Arg(0)

	switch command {

	case "check":
		checkCmd()
	case "install":
		installCmd()
	default:
		fallthrough
	case "run":
		runCmd(gracefulStop)
	}

}

func helpCmd() {
	_, err := os.Stderr.WriteString(
		`usage alaaarm <command> [<args>]
Commands:
	run            Run the bot (Default)
	check          Check the config file for missing fields
	install        Start the installation process
	version        Version of the server
`)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func checkCmd() {
	application := newApplication(log)

	err := application.LoadConfiguration(*configPath)
	if err != nil {
		log.WithError(err).Fatal("An error occured while reading the configuration: ")
	}
	log.Info("Config file seems correct")
	os.Exit(0)
}

func installCmd() {
	application := newApplication(log)

	log.Info("Writing default configuration")

	err := application.CreateConfiguration(*configPath)
	if err != nil {
		log.WithError(err).Fatal("There was an error writing the default configuration")
	}
	os.Exit(0)
}

func runCmd(quit chan os.Signal) {
	application := newApplication(log)

	err := application.LoadConfiguration(*configPath)
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

func setLogLevel() {
	switch *logLevel {
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
