// native/desktop/maplefile-cli/internal/usecase/medto/get.go
package medto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
)

// GetMeFromCloudUseCase defines the interface for getting user profile from cloud
type GetMeFromCloudUseCase interface {
	Execute(ctx context.Context) (*medto.MeResponseDTO, error)
}

// getMeFromCloudUseCase implements the GetMeFromCloudUseCase interface
type getMeFromCloudUseCase struct {
	logger     *zap.Logger
	repository medto.MeDTORepository
}

// NewGetMeFromCloudUseCase creates a new use case for getting user profile from cloud
func NewGetMeFromCloudUseCase(
	logger *zap.Logger,
	repository medto.MeDTORepository,
) GetMeFromCloudUseCase {
	logger = logger.Named("GetMeFromCloudUseCase")
	return &getMeFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute gets the current user's profile from cloud
func (uc *getMeFromCloudUseCase) Execute(ctx context.Context) (*medto.MeResponseDTO, error) {
	//
	// STEP 1: Validate preconditions (minimal for get operation)
	//

	uc.logger.Debug("Executing get me from cloud use case")

	//
	// STEP 2: Call repository to get user profile from cloud
	//

	response, err := uc.repository.GetMeFromCloud(ctx)
	if err != nil {
		uc.logger.Error("Failed to get user profile from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get user profile from the cloud", err)
	}

	//
	// STEP 3: Return response
	//

	uc.logger.Info("Successfully retrieved user profile from cloud",
		zap.String("email", response.Email),
		zap.String("name", response.Name))

	return response, nil
}
