# dsw (Do Something When…)

Created to overcome the limitations of tools like TriggerCMD which requires a cloud connection to execute local commands, has a clunky user interface, and imposes a very strict rate limit (1 action per minute), making fast automation or batch executions difficult. dsw offers a modern, lightweight, open-source alternative. It allows defining local actions that can be executed via CLI or a simple local HTTP API, without relying on the cloud. The goal is a simple, safe, and efficient tool to run pre-defined shell commands.

## Getting started

This project requires **Go version 1.25.4 or higher**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

1. **Installation**: Installs dsw in the system

    ```bash
    go install github.com/albertoboccolini/dsw@latest
    ```

2. Example usage:

    ```bash
    # Create a new action
    dsw create chromium "chromium"

    # Start the HTTP server for local API (daemon mode)
    dsw serve -d

    # Actions are defined in ~/.dsw/configuration.yaml and can be executed through the local HTTP API calls.
    curl -X POST http://localhost:8080/execute/chromium
    ```


## Commands

- `dsw create <name> <command>`: Create a single action
- `dsw create -f <file.yaml>`: Create actions from YAML file
- `dsw serve [-p 8080] [-d]`: Start HTTP API server (use -d for daemon mode)
- `dsw stop`: Stop daemon server
- `dsw boot enable [-p 8080]`: Enable automatic startup at boot (systemd user service)
- `dsw boot disable`: Disable automatic startup
- `dsw version`: Show version

## Limitations

Currently, dsw has the following limitations:

1. The HTTP server and actions work, but **integration with Alexa or other smart assistants is not yet supported**, since these platforms mainly rely on cloud services and have strict security checks.
2. No hot-reload for configuration; changes require a server restart.
3. No authentication or advanced concurrency control yet.
4. Boot service only works on Linux systems with systemd.

Recommended usage is **local deployment** with API calls triggered from shortcuts (e.g., iPhone + Siri) or via IFTTT.
