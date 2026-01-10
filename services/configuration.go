package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/albertoboccolini/dsw/models"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Actions map[string]models.Action `yaml:"actions" mapstructure:"actions"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Actions: make(map[string]models.Action),
	}
}

var actionNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (configuration *Configuration) GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".dsw")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create configuration directory: %w", err)
	}

	return filepath.Join(configDir, "configuration.yaml"), nil
}

func (configuration *Configuration) GetPIDPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".dsw", "dsw.pid"), nil
}

func (configuration *Configuration) Load() error {
	configPath, err := configuration.GetConfigPath()
	if err != nil {
		return err
	}

	yamlConfig := viper.New()
	yamlConfig.SetConfigFile(configPath)
	yamlConfig.SetConfigType("yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	if err := yamlConfig.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	if err := yamlConfig.Unmarshal(&configuration); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if configuration.Actions == nil {
		configuration.Actions = make(map[string]models.Action)
	}

	return nil
}

func (configuration *Configuration) Save() error {
	configPath, err := configuration.GetConfigPath()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"actions": configuration.Actions,
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, yamlData, 0600); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename configuration: %w", err)
	}

	return nil
}

func (configuration *Configuration) AddAction(name string, action models.Action) error {
	normalizedName := normalizeActionName(name)

	if !isValidActionName(normalizedName) {
		return fmt.Errorf("invalid action name: use only letters, numbers, dash and underscore")
	}

	configuration.Actions[normalizedName] = action
	return nil
}

func (configuration *Configuration) GetAction(name string) (models.Action, bool) {
	normalizedName := normalizeActionName(name)
	action, exists := configuration.Actions[normalizedName]
	return action, exists
}

func normalizeActionName(name string) string {
	return name
}

func isValidActionName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return actionNamePattern.MatchString(name)
}
