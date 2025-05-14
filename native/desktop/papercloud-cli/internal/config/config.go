// Package config provides a unified API for managing application configuration
// Location: monorepo/native/desktop/papercloud-cli/internal/config/config.go
package config

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

const (
	// AppName is the name of the application, used for configuration directories
	AppName = "papercloud-cli"
)

// Config holds all application configuration in a flat structure
type Config struct {
	// AppDirPath is the path to the directory where all files for this application are saved.
	AppDirPath string `json:"app_dir_path"`

	// CloudProviderAddress is the URI backend to make all calls to from this application.= for E2EE cloud operations.
	CloudProviderAddress string `json:"cloud_provider_address"`

	// Email is the unique registered email of the user whom successfully logged into the system.
	Email string `json:"email"`
}

// ConfigService defines the unified interface for all configuration operations
type ConfigService interface {
	GetCloudProviderAddress(ctx context.Context) (string, error)
	SetCloudProviderAddress(ctx context.Context, address string) error
	GetEmail(ctx context.Context) (string, error)
	SetEmail(ctx context.Context, email string) error
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string) (any, error)
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
		AppDirPath:           appConfigDir,
		Email:                "", // Leave blank because no user was authenticated.
	}
}

// Implementation of ConfigService methods

// getConfig is an internal method to get the current configuration
func (s *configService) getConfig(ctx context.Context) (*Config, error) {
	return s.repo.LoadConfig(ctx)
}

// saveConfig is an internal method to save the configuration
func (s *configService) saveConfig(ctx context.Context, config *Config) error {
	return s.repo.SaveConfig(ctx, config)
}

// GetCloudProviderAddress returns the cloud provider address
func (s *configService) GetCloudProviderAddress(ctx context.Context) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.CloudProviderAddress, nil
}

// SetCloudProviderAddress updates the cloud provider address
func (s *configService) SetCloudProviderAddress(ctx context.Context, address string) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.CloudProviderAddress = address
	return s.saveConfig(ctx, config)
}

// SetEmail updates the authenticated users email.
func (s *configService) SetEmail(ctx context.Context, email string) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.Email = email
	return s.saveConfig(ctx, config)
}

// GetEmail returns the authenticated users email.
func (s *configService) GetEmail(ctx context.Context) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.Email, nil
}

// Set saves a value to the config by key
func (s *configService) Set(ctx context.Context, key string, value any) error {
	// This method is a stub to implement the interface
	// It will be updated in a future version to save values to a key-value store
	return nil
}

// Get retrieves a value from the config by key
func (s *configService) Get(ctx context.Context, key string) (any, error) {
	// This method is a stub to implement the interface
	// It will be updated in a future version to retrieve values from a key-value store
	return nil, nil
}

// LevelDB support functions - updated to use app directory path

// NewLevelDBConfigurationProviderForUser returns a LevelDB configuration provider for users
func NewLevelDBConfigurationProviderForUser() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "users")
}

// NewLevelDBConfigurationProviderForCollection returns a LevelDB configuration provider for collections
func NewLevelDBConfigurationProviderForCollection() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "collections")
}

// Ensure our implementation satisfies the interface
var _ ConfigService = (*configService)(nil)
