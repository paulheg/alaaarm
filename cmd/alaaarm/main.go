package main

import (
	"flag"
	"log"
	"os"
)

var version = flag.Bool("version", false, "Show alaaarm version")
var configPath = flag.String("config", "./config.json", "Path to configuration file")

func main() {

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
		runCmd()
	}

}

func helpCmd() {
	_, err := os.Stderr.WriteString(
		`usage alaaarm <command> [<args>]
Commands:
	run            Run the bot (Default)
	check          Check the config file for missing fields
	install        Start the installation process
`)

	if err != nil {
		log.Fatal(err)
	}

}

func checkCmd() {

}

func installCmd() {

}

func runCmd() error {
	application := newApplication()

	err := application.LoadConfiguration("")
	if err != nil {
		log.Fatal("An error occured while reading the configuration: ", err)
	}

	err = application.Initialize()
	if err != nil {
		log.Fatal("An error occured while initializing the application: ", err)
	}

	application.Run()

	log.Print("Application finished")
	return err
}
