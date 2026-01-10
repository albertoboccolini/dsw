package main

import (
	"fmt"
	"os"

	"github.com/albertoboccolini/dsw/services"
)

func printUsage() {
	fmt.Println("DSW - Do Something When")
	fmt.Println("\nUsage:")
	fmt.Println("  dsw create <name> <command>     Create a single action")
	fmt.Println("  dsw create -f <file.yaml>       Create actions from YAML file")
	fmt.Println("  dsw serve [-p 8080] [-d]        Start HTTP API server")
	fmt.Println("  dsw stop                        Stop daemon server")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	configuration := services.NewConfiguration()
	err := configuration.Load()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	validator := services.NewValidator()
	daemon := services.NewDaemon()
	commandHandler := services.NewCommandHandler(configuration, validator, daemon)

	switch command {
	case "create":
		commandHandler.Create()
	case "serve":
		commandHandler.Serve()
	case "stop":
		commandHandler.ServerStop()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
