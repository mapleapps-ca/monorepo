// native/desktop/maplefile-cli/internal/service/me/get.go
package me

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
	uc_medto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/medto"
)

// GetOutput represents the result of getting user profile
type GetOutput struct {
	Profile *medto.MeResponseDTO `json:"profile"`
}

// GetMeService defines the interface for getting user profile
type GetMeService interface {
	Get(ctx context.Context) (*GetOutput, error)
}

// getMeService implements the GetMeService interface
type getMeService struct {
	logger                *zap.Logger
	getMeFromCloudUseCase uc_medto.GetMeFromCloudUseCase
}

// NewGetMeService creates a new service for getting user profile
func NewGetMeService(
	logger *zap.Logger,
	getMeFromCloudUseCase uc_medto.GetMeFromCloudUseCase,
) GetMeService {
	logger = logger.Named("Me.GetMeService")
	return &getMeService{
		logger:                logger,
		getMeFromCloudUseCase: getMeFromCloudUseCase,
	}
}

// Get retrieves the current user's profile from the cloud
func (s *getMeService) Get(ctx context.Context) (*GetOutput, error) {
	s.logger.Debug("Getting user profile")

	//
	// STEP 1: Get profile from cloud
	//

	profile, err := s.getMeFromCloudUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get user profile from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get user profile", err)
	}

	//
	// STEP 2: Return response
	//

	s.logger.Info("✅ Successfully retrieved user profile",
		zap.String("email", profile.Email),
		zap.String("name", profile.Name))

	return &GetOutput{
		Profile: profile,
	}, nil
}
