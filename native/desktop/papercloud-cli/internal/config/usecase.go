// config/usecase.go
package config

import (
	"context"
)

// configUseCase implements the ConfigUseCase interface
type configUseCase struct {
	repo ConfigRepository
}

// NewConfigUseCase creates a new instance of ConfigUseCase
func NewConfigUseCase(repo ConfigRepository) ConfigUseCase {
	return &configUseCase{
		repo: repo,
	}
}

// GetConfig returns the entire configuration
func (uc *configUseCase) GetConfig(ctx context.Context) (*Config, error) {
	return uc.repo.LoadConfig(ctx)
}

// GetCloudProviderAddress returns the cloud provider address
func (uc *configUseCase) GetCloudProviderAddress(ctx context.Context) (string, error) {
	config, err := uc.repo.LoadConfig(ctx)
	if err != nil {
		return "", err
	}
	return config.Application.CloudProviderAddress, nil
}

// SetCloudProviderAddress updates the cloud provider address
func (uc *configUseCase) SetCloudProviderAddress(ctx context.Context, address string) error {
	config, err := uc.repo.LoadConfig(ctx)
	if err != nil {
		return err
	}

	config.Application.CloudProviderAddress = address
	return uc.repo.SaveConfig(ctx, config)
}
