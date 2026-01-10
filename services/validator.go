package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (validator *Validator) ValidateCommand(commandPath string) error {
	if commandPath == "" {
		return fmt.Errorf("command cannot be empty")
	}

	resolvedPath, err := exec.LookPath(commandPath)
	if err != nil {
		return fmt.Errorf("command not found in PATH: %s", commandPath)
	}

	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return fmt.Errorf("cannot stat command: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("command is a directory: %s", resolvedPath)
	}

	filePermissions := fileInfo.Mode()
	if filePermissions&0111 == 0 {
		return fmt.Errorf("command is not executable: %s", resolvedPath)
	}

	return nil
}

func (validator *Validator) ParseCommandString(input string) (string, []string, error) {
	if input == "" {
		return "", nil, fmt.Errorf("command string is empty")
	}

	tokens, err := validator.splitCommand(input)
	if err != nil {
		return "", nil, err
	}

	if len(tokens) == 0 {
		return "", nil, fmt.Errorf("no command found")
	}

	command := tokens[0]
	args := []string{}
	if len(tokens) > 1 {
		args = tokens[1:]
	}

	if err := validator.ValidateCommand(command); err != nil {
		return "", nil, err
	}

	return command, args, nil
}

func (validator *Validator) splitCommand(input string) ([]string, error) {
	var tokens []string
	var currentToken strings.Builder

	insideQuotes := false
	activeQuote := rune(0)

	for position, character := range input {
		isQuoteCharacter := character == '"' || character == '\''
		isWhitespace := character == ' ' || character == '\t'

		if isQuoteCharacter {
			if !insideQuotes {
				insideQuotes = true
				activeQuote = character
			} else if character == activeQuote {
				insideQuotes = false
				activeQuote = 0
			} else {
				currentToken.WriteRune(character)
			}
			continue
		}

		if isWhitespace && !insideQuotes {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			continue
		}

		currentToken.WriteRune(character)

		isLastCharacter := position == len(input)-1
		if isLastCharacter && insideQuotes {
			return nil, fmt.Errorf("unclosed quote in command")
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens, nil
}
