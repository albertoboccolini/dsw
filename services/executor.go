package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/albertoboccolini/dsw/models"
)

const commandTimeout = 60 * time.Second

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (executor *Executor) Execute(action models.Action) models.ApiResponse {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	startTime := time.Now()
	result := executor.executeCommand(ctx, action)
	result.DurationMs = time.Since(startTime).Milliseconds()

	return result
}

func (executor *Executor) executeCommand(ctx context.Context, action models.Action) models.ApiResponse {
	fullCommand := action.Command
	if len(action.Args) > 0 {
		fullCommand = action.Command + " " + strings.Join(action.Args, " ")
	}

	command := exec.CommandContext(ctx, "sh", "-c", fullCommand)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer

	err := command.Run()

	combinedOutput := stdoutBuffer.String()
	if stderrBuffer.Len() > 0 {
		if len(combinedOutput) > 0 {
			combinedOutput += "\n"
		}

		combinedOutput += stderrBuffer.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return models.ApiResponse{
				Success: false,
				Output:  combinedOutput,
				Message: "Command timed out after 60 seconds",
			}
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			return models.ApiResponse{
				Success: false,
				Output:  combinedOutput,
				Message: fmt.Sprintf("Command failed with exit code %d", exitErr.ExitCode()),
			}
		}

		return models.ApiResponse{
			Success: false,
			Output:  combinedOutput,
			Message: fmt.Sprintf("Command error: %v", err),
		}
	}

	return models.ApiResponse{
		Success: true,
		Output:  combinedOutput,
		Message: "Command executed successfully",
	}
}
