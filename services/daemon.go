package services

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

type Daemon struct {
	configuration *Configuration
}

func NewDaemon() *Daemon {
	return &Daemon{}
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

	command := exec.Command(executable, "serve", "-p", strconv.Itoa(port))
	command.Stdout = nil
	command.Stderr = nil
	command.Stdin = nil

	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pid := command.Process.Pid
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0600); err != nil {
		command.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	command.Process.Release()

	fmt.Printf("Daemon started with PID %d on port %d\n", pid, port)
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
		os.Remove(pidPath)
		return fmt.Errorf("process not found: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		os.Remove(pidPath)
		return fmt.Errorf("failed to stop process: %w", err)
	}

	os.Remove(pidPath)
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
