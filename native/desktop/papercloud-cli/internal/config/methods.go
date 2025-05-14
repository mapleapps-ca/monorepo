// Package config provides a unified API for managing application configuration
// Location: monorepo/native/desktop/papercloud-cli/internal/config/config.go
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Implementation of ConfigService methods

// getConfig is an internal method to get the current configuration
func (s *configService) getConfig(ctx context.Context) (*Config, error) {
	return s.repo.LoadConfig(ctx)
}

// saveConfig is an internal method to save the configuration
func (s *configService) saveConfig(ctx context.Context, config *Config) error {
	return s.repo.SaveConfig(ctx, config)
}

func (s *configService) GetAppDirPath(ctx context.Context) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	return filepath.Join(configDir, AppName), nil
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

// Ensure our implementation satisfies the interface
var _ ConfigService = (*configService)(nil)
