// monorepo/native/desktop/papercloud-cli/config/config.go
package config

import "context"

// Config holds all application configuration
type Config struct {
	Application ApplicationConfig
}

// ApplicationConfig holds application-specific configuration
type ApplicationConfig struct {
	CloudProviderAddress string `json:"cloud_provider_address"`
}

// ConfigRepository defines the interface for loading and saving configuration
type ConfigRepository interface {
	// LoadConfig loads the configuration, returning defaults if file doesn't exist
	LoadConfig(ctx context.Context) (*Config, error)

	// SaveConfig saves the configuration to persistent storage
	SaveConfig(ctx context.Context, config *Config) error
}

// ConfigUseCase defines the business logic for working with configuration
type ConfigUseCase interface {
	// GetConfig returns the entire configuration
	GetConfig(ctx context.Context) (*Config, error)

	// GetCloudProviderAddress returns the cloud provider address
	GetCloudProviderAddress(ctx context.Context) (string, error)

	// SetCloudProviderAddress updates the cloud provider address
	SetCloudProviderAddress(ctx context.Context, address string) error
}

// GetDefaultConfig returns the default configuration values
func GetDefaultConfig() *Config {
	return &Config{
		Application: ApplicationConfig{
			CloudProviderAddress: "http://localhost:8000",
		},
	}
}
