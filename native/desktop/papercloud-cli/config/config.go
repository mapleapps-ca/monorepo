package config

// Config holds all application configuration
type Config struct {
	Application ApplicationConfig
}

// ApplicationConfig holds application-specific configuration
type ApplicationConfig struct {
	CloudProviderAddress string
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Application: ApplicationConfig{
			CloudProviderAddress: "TODO",
		},
	}
}
