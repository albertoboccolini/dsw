# dsw (Do Something When…)

Created to overcome the limitations of tools like TriggerCMD which requires a cloud connection to execute local commands, has a clunky user interface, and imposes a very strict rate limit (1 action per minute), making fast automation or batch executions difficult. dsw offers a modern, lightweight, open-source alternative. It allows defining local actions that can be executed via CLI or a simple local HTTP API, without relying on the cloud. The goal is a simple, safe, and efficient tool to run pre-defined shell commands.

## Getting started

This project requires **Go version 1.25.4 or higher**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

Since is still under development, it should be cloned and installed using Go:

```bash
git clone https://github.com/albertoboccolini/dsw
cd dsw
go install
```

Example usage:

```bash
# Create a new action
dsw create chromium "chromium"

# Start the HTTP server for local API
dsw serve -d

# Actions are defined in ~/.dsw/config.yaml and can be executed through the local HTTP API calls.
curl -X POST http://localhost:8080/execute/chromium
```


## Limitations

Currently, dsw has the following limitations:

1. The HTTP server and actions work, but **integration with Alexa or other smart assistants is not yet supported**, since these platforms mainly rely on cloud services and have strict security checks.
2. No hot-reload for configuration; changes require a server restart.
3. No authentication or advanced concurrency control yet.

Recommended usage is **local deployment** with API calls triggered from shortcuts (e.g., iPhone + Siri) or via IFTTT.
