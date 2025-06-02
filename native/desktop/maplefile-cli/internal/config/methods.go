// Package config provides a unified API for managing application configuration
// Location: monorepo/native/desktop/maplefile-cli/internal/config/config.go
package config

import (
	"context"
	"time"
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

// GetAppDataDirPath returns the proper application data directory path
func (s *configService) GetAppDataDirPath(ctx context.Context) (string, error) {
	return GetUserDataDir(AppName)
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

// SetLoggedInUserEmail updates the authenticated users email.
func (s *configService) SetLoggedInUserCredentials(
	ctx context.Context,
	email string,
	accessToken string,
	accessTokenExpiryTime *time.Time,
	refreshToken string,
	refreshTokenExpiryTime *time.Time,
) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	config.Credentials = &Credentials{
		Email:                  email,
		AccessToken:            accessToken,
		AccessTokenExpiryTime:  accessTokenExpiryTime,
		RefreshToken:           refreshToken,
		RefreshTokenExpiryTime: refreshTokenExpiryTime,
	}
	return s.saveConfig(ctx, config)
}

// GetLoggedInUserCredentials returns the authenticated user's credentials.
func (s *configService) GetLoggedInUserCredentials(ctx context.Context) (*Credentials, error) {
	config, err := s.getConfig(ctx)
	if err != nil {
		return nil, err
	}
	return config.Credentials, nil
}

func (s *configService) ClearLoggedInUserCredentials(ctx context.Context) error {
	config, err := s.getConfig(ctx)
	if err != nil {
		return err
	}

	// Clear credentials by setting them to empty values
	config.Credentials = &Credentials{
		Email:                  "",
		AccessToken:            "",
		AccessTokenExpiryTime:  nil,
		RefreshToken:           "",
		RefreshTokenExpiryTime: nil,
	}

	return s.saveConfig(ctx, config)
}

// Ensure our implementation satisfies the interface
var _ ConfigService = (*configService)(nil)
