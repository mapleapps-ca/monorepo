// native/desktop/maplefile-cli/internal/service/me/update.go
package me

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
	uc_medto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/medto"
)

// UpdateInput represents the input for updating user profile
type UpdateInput struct {
	Email                                          string `json:"email"`
	FirstName                                      string `json:"first_name"`
	LastName                                       string `json:"last_name"`
	Phone                                          string `json:"phone,omitempty"`
	Country                                        string `json:"country,omitempty"`
	Region                                         string `json:"region,omitempty"`
	Timezone                                       string `json:"timezone"`
	AgreePromotions                                bool   `json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
}

// UpdateOutput represents the result of updating user profile
type UpdateOutput struct {
	Profile *medto.MeResponseDTO `json:"profile"`
}

// UpdateMeService defines the interface for updating user profile
type UpdateMeService interface {
	Update(ctx context.Context, input *UpdateInput) (*UpdateOutput, error)
}

// updateMeService implements the UpdateMeService interface
type updateMeService struct {
	logger                 *zap.Logger
	updateMeInCloudUseCase uc_medto.UpdateMeInCloudUseCase
}

// NewUpdateMeService creates a new service for updating user profile
func NewUpdateMeService(
	logger *zap.Logger,
	updateMeInCloudUseCase uc_medto.UpdateMeInCloudUseCase,
) UpdateMeService {
	logger = logger.Named("Me.UpdateMeService")
	return &updateMeService{
		logger:                 logger,
		updateMeInCloudUseCase: updateMeInCloudUseCase,
	}
}

// Update updates the current user's profile in the cloud
func (s *updateMeService) Update(ctx context.Context, input *UpdateInput) (*UpdateOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	s.logger.Debug("Updating user profile",
		zap.String("email", input.Email),
		zap.String("first_name", input.FirstName),
		zap.String("last_name", input.LastName))

	//
	// STEP 2: Convert to DTO request
	//

	request := &medto.UpdateMeRequestDTO{
		Email:           input.Email,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Phone:           input.Phone,
		Country:         input.Country,
		Region:          input.Region,
		Timezone:        input.Timezone,
		AgreePromotions: input.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: input.AgreeToTrackingAcrossThirdPartyAppsAndServices,
	}

	//
	// STEP 3: Update profile in cloud
	//

	profile, err := s.updateMeInCloudUseCase.Execute(ctx, request)
	if err != nil {
		s.logger.Error("❌ Failed to update user profile in cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to update user profile", err)
	}

	//
	// STEP 4: Return response
	//

	s.logger.Info("✅ Successfully updated user profile",
		zap.String("email", profile.Email),
		zap.String("name", profile.Name))

	return &UpdateOutput{
		Profile: profile,
	}, nil
}
