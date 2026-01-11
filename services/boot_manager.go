package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const bootDisableMessage = "Boot disabled successfully"
const bootEnableMessage = "Boot enabled successfully"

const serviceTemplate = `[Unit]
Description=DSW Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s serve -p %d
Restart=on-failure
RestartSec=10
Environment="PATH=%s"
Environment="HOME=%s"
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`

type BootManager struct {
	configuration *Configuration
}

func NewBootManager(configuration *Configuration) *BootManager {
	return &BootManager{configuration: configuration}
}

func (bootManager *BootManager) getServicePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	return filepath.Join(home, ".config", "systemd", "user", "dsw.service")
}

func (bootManager *BootManager) getExecutablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(executable)
}

func (bootManager *BootManager) runSystemctl(args ...string) error {
	fullArgs := append([]string{"--user"}, args...)
	command := exec.Command("systemctl", fullArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (bootManager *BootManager) EnableBootService(port int) error {
	execPath, err := bootManager.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		pathEnv = "/usr/local/bin:/usr/bin:/bin"
	}

	serviceContent := fmt.Sprintf(serviceTemplate,
		execPath,
		port,
		pathEnv,
		home,
	)

	serviceDir := filepath.Dir(bootManager.getServicePath())
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	if err := os.WriteFile(bootManager.getServicePath(), []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	if err := bootManager.runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := bootManager.runSystemctl("enable", "dsw.service"); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	fmt.Println(bootEnableMessage)
	return nil
}

func (bootManager *BootManager) DisableBootService() error {
	servicePath := bootManager.getServicePath()

	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return fmt.Errorf("boot is not configured")
	}

	if err := bootManager.runSystemctl("disable", "dsw.service"); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	if err := bootManager.runSystemctl("stop", "dsw.service"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to stop service: %v\n", err)
	}

	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	if err := bootManager.runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	fmt.Println(bootDisableMessage)
	return nil
}

func (bootManager *BootManager) IsBootServiceEnabled() bool {
	servicePath := bootManager.getServicePath()
	_, err := os.Stat(servicePath)
	return err == nil
}
