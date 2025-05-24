// Package config provides a unified API for managing application configuration
// Location: monorepo/native/desktop/maplefile-cli/internal/config/config.go
package config

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
)

const (
	// AppName is the name of the application, used for configuration directories
	AppName = "maplefile-cli"
)

// Config holds all application configuration in a flat structure
type Config struct {
	// CloudProviderAddress is the URI backend to make all calls to from this application.= for E2EE cloud operations.
	CloudProviderAddress string       `json:"cloud_provider_address"`
	Credentials          *Credentials `json:"credentials"`
}

// Credentials holds all user credentials for authentication and authorization. Values are decrypted for convenience purposes as we assume threat actor cannot access the decrypted values on the user's device.
type Credentials struct {
	// Email is the unique registered email of the user whom successfully logged into the system.
	Email                  string     `json:"email"`
	AccessToken            string     `json:"access_token"`
	AccessTokenExpiryTime  *time.Time `json:"access_token_expiry_time"`
	RefreshToken           string     `json:"refresh_token"`
	RefreshTokenExpiryTime *time.Time `json:"refresh_token_expiry_time"`
}

// ConfigService defines the unified interface for all configuration operations
type ConfigService interface {
	GetAppDataDirPath(ctx context.Context) (string, error)
	GetCloudProviderAddress(ctx context.Context) (string, error)
	SetCloudProviderAddress(ctx context.Context, address string) error
	GetLoggedInUserCredentials(ctx context.Context) (*Credentials, error)
	SetLoggedInUserCredentials(
		ctx context.Context,
		email string,
		accessToken string,
		accessTokenExpiryTime *time.Time,
		refreshToken string,
		refreshTokenExpiryTime *time.Time,
	) error
}

// repository defines the interface for loading and saving configuration
type repository interface {
	// LoadConfig loads the configuration, returning defaults if file doesn't exist
	LoadConfig(ctx context.Context) (*Config, error)

	// SaveConfig saves the configuration to persistent storage
	SaveConfig(ctx context.Context, config *Config) error
}

// configService implements the ConfigService interface
type configService struct {
	repo repository
}

// fileRepository implements the repository interface with file-based storage
type fileRepository struct {
	configPath string
	appName    string
}

// New creates a new configuration service with default settings
func New() (ConfigService, error) {
	repo, err := newFileRepository(AppName)
	if err != nil {
		return nil, err
	}

	return &configService{
		repo: repo,
	}, nil
}

// NewForTesting creates a configuration service with the specified repository (for testing)
func NewForTesting(repo repository) ConfigService {
	return &configService{
		repo: repo,
	}
}

// Module returns fx options to register the configuration service
func Module() fx.Option {
	return fx.Provide(
		New,
	)
}

// newFileRepository creates a new instance of repository
func newFileRepository(appName string) (repository, error) {
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

	return &fileRepository{
		configPath: configPath,
		appName:    appName,
	}, nil
}

// LoadConfig loads the configuration from file, or returns defaults if file doesn't exist
func (r *fileRepository) LoadConfig(ctx context.Context) (*Config, error) {
	// Check if the config file exists
	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		defaults := getDefaultConfig()

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
func (r *fileRepository) SaveConfig(ctx context.Context, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.configPath, data, 0644)
}

// getDefaultConfig returns the default configuration values
func getDefaultConfig() *Config {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user home directory with error: %v\n", err)
	}

	// Create app-specific config directory
	appConfigDir := filepath.Join(configDir, AppName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		log.Fatalf("Failed getting user home directory with error: %v\n", err)
	}

	return &Config{
		CloudProviderAddress: "http://localhost:8000",
		Credentials: &Credentials{
			Email:                  "",  // Leave blank because no user was authenticated.
			AccessToken:            "",  // Leave blank because no user was authenticated.
			AccessTokenExpiryTime:  nil, // Leave blank because no user was authenticated.
			RefreshToken:           "",  // Leave blank because no user was authenticated.
			RefreshTokenExpiryTime: nil, // Leave blank because no user was authenticated.
		},
	}
}
