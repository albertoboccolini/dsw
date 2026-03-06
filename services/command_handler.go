package services

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/albertoboccolini/dsw/models"
	"github.com/spf13/viper"
)

type CommandHandler struct {
	configuration *Configuration
	validator     *Validator
	daemon        *Daemon
	utils         *Utils
}

func NewCommandHandler(
	configuration *Configuration,
	validator *Validator,
	daemon *Daemon,
	utils *Utils,
) *CommandHandler {
	return &CommandHandler{
		configuration: configuration,
		validator:     validator,
		daemon:        daemon,
		utils:         utils,
	}
}

func (commandHandler *CommandHandler) singleCreate(actionName, commandString string) {
	command, args, err := commandHandler.validator.ParseCommandString(commandString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid command: %v\n", err)
		os.Exit(1)
	}

	action := models.Action{
		Command: command,
		Args:    args,
	}

	if err := commandHandler.configuration.AddAction(actionName, action); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to add action: %v\n", err)
		os.Exit(1)
	}

	if err := commandHandler.configuration.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Action '%s' created successfully\n", actionName)
	fmt.Printf("  Command: %s\n", command)
	fmt.Printf("  Args: %v\n", args)
}

func (commandHandler *CommandHandler) batchCreate(filePath string) {
	yamlConfig := viper.New()
	yamlConfig.SetConfigFile(filePath)
	yamlConfig.SetConfigType("yaml")

	if err := yamlConfig.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read configuration file: %v\n", err)
		os.Exit(1)
	}

	var batchConfig Configuration
	if err := yamlConfig.Unmarshal(&batchConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse configuration file: %v\n", err)
		os.Exit(1)
	}

	addedCount := 0
	for name, action := range batchConfig.Actions {
		if err := commandHandler.validator.ValidateCommand(action.Command); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping action '%s': %v\n", name, err)
			continue
		}

		if err := commandHandler.configuration.AddAction(name, action); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to add action '%s': %v\n", name, err)
			continue
		}

		addedCount++
	}

	if err := commandHandler.configuration.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added %d action(s) from %s\n", addedCount, filePath)
}

func (commandHandler *CommandHandler) Create() {
	createFlags := flag.NewFlagSet("create", flag.ExitOnError)
	configFile := createFlags.String("f", "", "YAML file with actions to add")
	err := createFlags.Parse(os.Args[2:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse flags: %v\n", err)
		os.Exit(1)
	}

	if *configFile != "" {
		commandHandler.batchCreate(*configFile)
		return
	}

	if createFlags.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: dsw create <name> <command>")
		os.Exit(1)
	}

	actionName := createFlags.Arg(0)
	commandString := createFlags.Arg(1)

	commandHandler.singleCreate(actionName, commandString)
}

func (commandHandler *CommandHandler) ListActions() {
	if len(commandHandler.configuration.Actions) == 0 {
		fmt.Println("No actions configured")
		return
	}

	for name, action := range commandHandler.configuration.Actions {
		fmt.Printf("%s: %s", name, action.Command)
		if len(action.Args) > 0 {
			fmt.Printf(" %v", action.Args)
		}

		fmt.Println()
	}
}

func (commandHandler *CommandHandler) Serve() {
	serveFlags := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveFlags.Int("p", models.DEFAULT_PORT, "Port to listen on")
	daemonMode := serveFlags.Bool("d", false, "Run in daemon mode")
	err := serveFlags.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse flags: %v\n", err)
		os.Exit(1)
	}

	if *daemonMode {
		if err := commandHandler.daemon.StartDaemon(*port); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return
	}

	if len(commandHandler.configuration.Actions) == 0 {
		slog.Warn("no actions configured")
	}

	serverHandler := NewServerHandler(commandHandler.configuration, *port)
	if err := serverHandler.Server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: server failed: %v\n", err)
		os.Exit(1)
	}
}

func (commandHandler *CommandHandler) ServerStop() {
	if err := commandHandler.daemon.StopDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (commandHandler *CommandHandler) Status() {
	bootManager := NewBootManager(commandHandler.configuration)

	status := "not running"
	pid := "N/A"
	port := "N/A"

	if commandHandler.daemon.IsRunning() {
		status = "running"
		pidPath, _ := commandHandler.configuration.GetPIDPath()
		pidData, err := os.ReadFile(pidPath)
		if err == nil {
			pid = strings.TrimSpace(string(pidData))
			port = commandHandler.utils.ExtractPortFromCmdline(pid)
		}
	}

	fmt.Printf("Status: %s\n", status)
	bootEnabled := bootManager.IsBootServiceEnabled()
	bootStatus := "disabled"
	if bootEnabled {
		bootStatus = "enabled"
	}

	fmt.Printf("Boot: %s\n", bootStatus)
	fmt.Printf("Port: %s\n", port)
	fmt.Printf("PID: %s\n", pid)
}

func (commandHandler *CommandHandler) HandleBoot() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: boot subcommand required (enable|disable)")
		os.Exit(1)
	}

	bootCommand := os.Args[2]
	bootManager := NewBootManager(commandHandler.configuration)

	switch bootCommand {
	case "enable":
		bootFlags := flag.NewFlagSet("boot enable", flag.ExitOnError)
		port := bootFlags.Int("p", models.DEFAULT_PORT, "Port to listen on")
		if err := bootFlags.Parse(os.Args[3:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse flags: %v\n", err)
			os.Exit(1)
		}

		if err := bootManager.EnableBootService(*port); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "disable":
		if err := bootManager.DisableBootService(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown boot command: %s\n", bootCommand)
		os.Exit(1)
	}
}
