// monorepo/native/desktop/papercloud-cli/config/config.go
package config

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/keys"
)

// Config holds all application configuration
type Config struct {
	Application ApplicationConfig
	Account     AccountConfig
}

// ApplicationConfig holds application-specific configuration
type ApplicationConfig struct {
	CloudProviderAddress string `json:"cloud_provider_address"`
}

type AccountConfig struct {
	Email                             string                                 `json:"email" bson:"email"`
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
