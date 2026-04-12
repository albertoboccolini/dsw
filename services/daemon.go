package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/albertoboccolini/dsw/models"
)

type Daemon struct {
	configuration *Configuration
}

func NewDaemon(configuration *Configuration) *Daemon {
	return &Daemon{
		configuration: configuration,
	}
}

func (daemon *Daemon) getLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, models.CONFIGURATION_FOLDER)
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	return filepath.Join(logDir, "dsw.log"), nil
}

func (daemon *Daemon) StartDaemon(port int) error {
	pidPath, err := daemon.configuration.GetPIDPath()
	if err != nil {
		return err
	}

	if daemon.IsRunning() {
		return fmt.Errorf("daemon already running")
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	logPath, err := daemon.getLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	defer func() {
		if closeErr := logFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close log file: %v\n", closeErr)
		}
	}()

	command := exec.Command(executable, "serve", "-p", strconv.Itoa(port))
	command.Stdout = logFile
	command.Stderr = logFile
	command.Stdin = nil
	command.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pid := command.Process.Pid
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0o600); err != nil {
		if err := command.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process after PID file write failure: %v", err)
		}

		return fmt.Errorf("failed to write PID file: %w", err)
	}

	if err := command.Process.Release(); err != nil {
		return fmt.Errorf("failed to release process: %v", err)
	}

	fmt.Printf("Daemon started with PID %d on port %d\n", pid, port)
	fmt.Printf("Logs: %s\n", logPath)
	return nil
}

func (daemon *Daemon) StopDaemon() error {
	pidPath, err := daemon.configuration.GetPIDPath()
	if err != nil {
		return err
	}

	if !daemon.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		err = os.Remove(pidPath)
		return fmt.Errorf("process not found: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		err = os.Remove(pidPath)
		return fmt.Errorf("failed to stop process: %w", err)
	}

	err = os.Remove(pidPath)
	if err != nil {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	fmt.Printf("Daemon stopped (PID %d)\n", pid)
	return nil
}

func (daemon *Daemon) IsRunning() bool {
	pidPath, err := daemon.configuration.GetPIDPath()
	if err != nil {
		return false
	}

	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
