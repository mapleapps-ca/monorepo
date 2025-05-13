// monorepo/native/desktop/papercloud-cli/config/filerepository.go
package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// FileConfigRepository implements the ConfigRepository interface
// storing configuration in a file in the appropriate location based on OS
type FileConfigRepository struct {
	configPath string
	appName    string
}

// NewFileConfigRepository creates a new instance of ConfigRepository
func NewFileConfigRepository(appName string) (ConfigRepository, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// Create app-specific config directory
	appConfigDir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(appConfigDir, "config.json")

	return &FileConfigRepository{
		configPath: configPath,
		appName:    appName,
	}, nil
}

// LoadConfig loads the configuration from file, or returns defaults if file doesn't exist
func (r *FileConfigRepository) LoadConfig(ctx context.Context) (*Config, error) {
	// Check if the config file exists
	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		defaults := GetDefaultConfig()

		// Save the defaults for future use
		if err := r.SaveConfig(ctx, defaults); err != nil {
			return nil, err
		}

		return defaults, nil
	}

	// Read config from file
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func (r *FileConfigRepository) SaveConfig(ctx context.Context, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.configPath, data, 0644)
}
