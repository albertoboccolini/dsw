package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/overthinkinglabs/dsw/src/models"
	"github.com/overthinkinglabs/dsw/src/services"
)

func printUsage() {
	fmt.Println("DSW - Do Something When")
	fmt.Println("\nUsage:")
	fmt.Println("  dsw create <name> <command>     Create a single action")
	fmt.Println("  dsw create -f <file.yaml>       Create actions from YAML file")
	fmt.Println("  dsw ls                          List all available actions")
	fmt.Println("  dsw run <name>                  Execute an action")
	fmt.Println("  dsw serve [-p 8080] [-d]        Start HTTP API server")
	fmt.Println("  dsw stop                        Stop daemon server")
	fmt.Println("  dsw status                      Show daemon status")
	fmt.Println("  dsw boot enable [-p 8080]       Enable boot service")
	fmt.Println("  dsw boot disable                Disable boot service")
	fmt.Println("  dsw version                     Show version")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	configuration := services.NewConfiguration()
	err := configuration.Load()
	var version bool
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.BoolVar(&version, "v", false, "Show version information (shorthand)")
	flag.Parse()

	if version {
		fmt.Printf("v%s\n", models.VERSION)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	validator := services.NewValidator()
	daemon := services.NewDaemon(configuration)
	utils := services.NewUtils()
	commandHandler := services.NewCommandHandler(configuration, validator, daemon, utils)

	switch command {
	case "create":
		commandHandler.Create()
	case "ls":
		commandHandler.ListActions()
	case "run":
		commandHandler.Run()
	case "serve":
		commandHandler.Serve()
	case "stop":
		commandHandler.ServerStop()
	case "status":
		commandHandler.Status()
	case "boot":
		commandHandler.HandleBoot()
	case "version":
		fmt.Printf("v%s\n", models.VERSION)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
