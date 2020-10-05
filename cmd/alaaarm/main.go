package main

import (
	"flag"
	"log"
	"os"
)

var version = flag.Bool("version", false, "Show alaaarm version")
var configPath = flag.String("config", "./config/config.json", "Path to configuration file")

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
	version        Version of the server
`)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func checkCmd() {
	application := newApplication()

	err := application.LoadConfiguration(*configPath)
	if err != nil {
		log.Fatal("An error occured while reading the configuration: ", err)
	}
	log.Println("Config file seems correct")
	os.Exit(0)
}

func installCmd() {
	application := newApplication()

	log.Println("Writing default configuration")

	err := application.CreateConfiguration(*configPath)
	if err != nil {
		log.Fatalf("There was an error writing the default configuration: %s", err.Error())
	}
	os.Exit(0)
}

func runCmd() {
	application := newApplication()

	err := application.LoadConfiguration(*configPath)
	if err != nil {
		log.Fatal("An error occured while reading the configuration: ", err)
	}

	err = application.Initialize()
	if err != nil {
		log.Fatal("An error occured while initializing the application: ", err)
	}

	application.Run()

	log.Print("Application finished")
}
