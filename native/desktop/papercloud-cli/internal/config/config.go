// Package config provides a unified API for managing application configuration
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

const (
	// AppName is the name of the application, used for configuration directories
	AppName = "papercloud-cli"
)

// Config holds all application configuration in a flat structure
type Config struct {
	CloudProviderAddress string `json:"cloud_provider_address"`

	FirstName                                      string `json:"first_name"`
	LastName                                       string `json:"last_name"`
	Email                                          string `json:"email"`
	Phone                                          string `json:"phone,omitempty"`
	Country                                        string `json:"country,omitempty"`
	CountryOther                                   string `json:"country_other,omitempty"`
	Timezone                                       string `bson:"timezone" json:"timezone"`
	AgreeTermsOfService                            bool   `json:"agree_terms_of_service,omitempty"`
	AgreePromotions                                bool   `json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`

	PasswordSalt                      []byte                                 `json:"password_salt" bson:"password_salt"`
	EncryptedMasterKey                keys.EncryptedMasterKey                `json:"encrypted_master_key" bson:"encrypted_master_key"`
	PublicKey                         keys.PublicKey                         `json:"public_key" bson:"public_key"`
	EncryptedPrivateKey               keys.EncryptedPrivateKey               `json:"encrypted_private_key" bson:"encrypted_private_key"`
	EncryptedRecoveryKey              keys.EncryptedRecoveryKey              `json:"encrypted_recovery_key" bson:"encrypted_recovery_key"`
	MasterKeyEncryptedWithRecoveryKey keys.MasterKeyEncryptedWithRecoveryKey `json:"master_key_encrypted_with_recovery_key" bson:"master_key_encrypted_with_recovery_key"`
	VerificationID                    string                                 `json:"verificationID"`
	AccessToken                       string                                 `json:"access_token"`
	AccessTokenExpiryTime             time.Time                              `json:"access_token_expiry_time"`
	RefreshToken                      string                                 `json:"refresh_token"`
	RefreshTokenExpiryTime            time.Time                              `json:"refresh_token_expiry_time"`
}

// ConfigService defines the unified interface for all configuration operations
type ConfigService interface {
	// Application settings
	GetCloudProviderAddress(ctx context.Context) (string, error)
	SetCloudProviderAddress(ctx context.Context, address string) error

	// Account settings
	GetEmail(ctx context.Context) (string, error)
	SetEmail(ctx context.Context, email string) error
	GetAccessToken(ctx context.Context) (string, error)
	SetAccessToken(ctx context.Context, token string) error
	SetAccessTokenWithExpiry(ctx context.Context, token string, expiryTime time.Time) error
	GetRefreshToken(ctx context.Context) (string, error)
	SetRefreshToken(ctx context.Context, token string) error
	SetRefreshTokenWithExpiry(ctx context.Context, token string, expiryTime time.Time) error
	GetEncryptedMasterKey(ctx context.Context) (keys.EncryptedMasterKey, error)
	SetEncryptedMasterKey(ctx context.Context, encryptedMasterKey keys.EncryptedMasterKey) error

	// Generic access
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
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
	return &Config{
		CloudProviderAddress: "http://localhost:8000",
		// All other fields default to zero values
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

// GetEmail returns the account email
func (s *configService) GetEmail(ctx context.Context) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.Email, nil
}

// SetEmail updates the account email
func (s *configService) SetEmail(ctx context.Context, email string) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.Email = email
	return s.saveConfig(ctx, config)
}

// GetAccessToken returns the account access token
func (s *configService) GetAccessToken(ctx context.Context) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.AccessToken, nil
}

// SetAccessToken updates the account access token
func (s *configService) SetAccessToken(ctx context.Context, token string) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.AccessToken = token
	return s.saveConfig(ctx, config)
}

// SetAccessTokenWithExpiry updates the access token and its expiry time
func (s *configService) SetAccessTokenWithExpiry(ctx context.Context, token string, expiryTime time.Time) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.AccessToken = token
	config.AccessTokenExpiryTime = expiryTime
	return s.saveConfig(ctx, config)
}

// GetRefreshToken returns the account refresh token
func (s *configService) GetRefreshToken(ctx context.Context) (string, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.RefreshToken, nil
}

// SetRefreshToken updates the account refresh token
func (s *configService) SetRefreshToken(ctx context.Context, token string) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.RefreshToken = token
	return s.saveConfig(ctx, config)
}

// SetRefreshTokenWithExpiry updates the refresh token and its expiry time
func (s *configService) SetRefreshTokenWithExpiry(ctx context.Context, token string, expiryTime time.Time) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.RefreshToken = token
	config.RefreshTokenExpiryTime = expiryTime
	return s.saveConfig(ctx, config)
}

// SetEncryptedMasterKey updates the encrypted master key
func (s *configService) SetEncryptedMasterKey(ctx context.Context, encryptedMasterKey keys.EncryptedMasterKey) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.EncryptedMasterKey = encryptedMasterKey
	return s.saveConfig(ctx, config)
}

// GetEncryptedMasterKey returns the encrypted master key
func (s *configService) GetEncryptedMasterKey(ctx context.Context) (keys.EncryptedMasterKey, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return keys.EncryptedMasterKey{}, err
	}
	return config.EncryptedMasterKey, nil
}

// Get retrieves a specific configuration value by key
func (s *configService) Get(ctx context.Context, key string) (interface{}, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return nil, err
	}

	switch key {
	case "cloud_provider_address":
		return config.CloudProviderAddress, nil
	case "email":
		return config.Email, nil
	case "access_token":
		return config.AccessToken, nil
	case "refresh_token":
		return config.RefreshToken, nil
	case "verification_id":
		return config.VerificationID, nil
	default:
		return nil, fmt.Errorf("unknown configuration key: %s", key)
	}
}

// Set updates a specific configuration value by key
func (s *configService) Set(ctx context.Context, key string, value interface{}) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	switch key {
	case "cloud_provider_address":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.CloudProviderAddress = strValue
	case "email":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.Email = strValue
	case "first_name":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.FirstName = strValue
	case "last_name":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.LastName = strValue
	case "password_salt":
		byteValue, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("value for %s must be a byte slice", key)
		}
		config.PasswordSalt = byteValue
	case "access_token":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.AccessToken = strValue
	case "refresh_token":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.RefreshToken = strValue
	case "verification_id":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", key)
		}
		config.VerificationID = strValue
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return s.saveConfig(ctx, config)
}

// LevelDB support functions - exported to maintain compatibility

// NewLevelDBConfigurationProviderForUser returns a LevelDB configuration provider for users
func NewLevelDBConfigurationProviderForUser() leveldb.LevelDBConfigurationProvider {
	return leveldb.NewLevelDBConfigurationProvider("./", "users")
}

// NewLevelDBConfigurationProviderForCollection returns a LevelDB configuration provider for collections
func NewLevelDBConfigurationProviderForCollection() leveldb.LevelDBConfigurationProvider {
	return leveldb.NewLevelDBConfigurationProvider("./", "collections")
}

// Ensure our implementation satisfies the interface
var _ ConfigService = (*configService)(nil)
