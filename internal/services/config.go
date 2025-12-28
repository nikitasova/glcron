package services

import (
	"encoding/json"
	"fmt"
	"glcron/internal/models"
	"os"
	"path/filepath"
)

// ConfigServiceInterface defines the interface for config storage
type ConfigServiceInterface interface {
	Load() (*models.ConfigFile, error)
	Save(configFile *models.ConfigFile) error
	GetConfigPath() string
	AddConfig(config models.Config) error
	UpdateConfig(index int, config models.Config) error
	DeleteConfig(index int) error
	GetConfigs() []models.Config
}

// ConfigService handles configuration file operations
type ConfigService struct {
	configPath string
	configFile *models.ConfigFile
}

// NewConfigService creates a new ConfigService
func NewConfigService() ConfigServiceInterface {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		// Fallback to current directory
		configDir = "."
	}

	return &ConfigService{
		configPath: filepath.Join(configDir, "glcron.json"),
		configFile: &models.ConfigFile{
			Configs: []models.Config{},
		},
	}
}

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	// Use XDG_CONFIG_HOME if set
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		dir := filepath.Join(xdgConfig, "glcron")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}
		return dir, nil
	}

	// Use ~/.config/glcron
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(homeDir, ".config", "glcron")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return dir, nil
}

// GetConfigPath returns the config file path
func (c *ConfigService) GetConfigPath() string {
	return c.configPath
}

// Load loads the configuration file
func (c *ConfigService) Load() (*models.ConfigFile, error) {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			c.configFile = &models.ConfigFile{
				Configs: []models.Config{},
			}
			return c.configFile, nil
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, c.configFile); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return c.configFile, nil
}

// Save saves the configuration file
func (c *ConfigService) Save(configFile *models.ConfigFile) error {
	c.configFile = configFile

	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(c.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// AddConfig adds a new configuration
func (c *ConfigService) AddConfig(config models.Config) error {
	c.configFile.Configs = append(c.configFile.Configs, config)
	return c.Save(c.configFile)
}

// UpdateConfig updates an existing configuration
func (c *ConfigService) UpdateConfig(index int, config models.Config) error {
	if index < 0 || index >= len(c.configFile.Configs) {
		return fmt.Errorf("invalid config index: %d", index)
	}

	c.configFile.Configs[index] = config
	return c.Save(c.configFile)
}

// DeleteConfig deletes a configuration
func (c *ConfigService) DeleteConfig(index int) error {
	if index < 0 || index >= len(c.configFile.Configs) {
		return fmt.Errorf("invalid config index: %d", index)
	}

	c.configFile.Configs = append(c.configFile.Configs[:index], c.configFile.Configs[index+1:]...)
	return c.Save(c.configFile)
}

// GetConfigs returns all configurations
func (c *ConfigService) GetConfigs() []models.Config {
	return c.configFile.Configs
}
