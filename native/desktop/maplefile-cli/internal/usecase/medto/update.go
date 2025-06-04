// native/desktop/maplefile-cli/internal/usecase/medto/update.go
package medto

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// UpdateMeInCloudUseCase defines the interface for updating user profile in cloud
type UpdateMeInCloudUseCase interface {
	Execute(ctx context.Context, request *medto.UpdateMeRequestDTO) (*medto.MeResponseDTO, error)
}

// updateMeInCloudUseCase implements the UpdateMeInCloudUseCase interface
type updateMeInCloudUseCase struct {
	logger     *zap.Logger
	repository medto.MeDTORepository
}

// NewUpdateMeInCloudUseCase creates a new use case for updating user profile in cloud
func NewUpdateMeInCloudUseCase(
	logger *zap.Logger,
	repository medto.MeDTORepository,
) UpdateMeInCloudUseCase {
	logger = logger.Named("UpdateMeInCloudUseCase")
	return &updateMeInCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute updates the current user's profile in cloud
func (uc *updateMeInCloudUseCase) Execute(ctx context.Context, request *medto.UpdateMeRequestDTO) (*medto.MeResponseDTO, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if request == nil {
		e["request"] = "Request is required"
	} else {
		// Sanitization
		request.Email = strings.ToLower(request.Email) // Ensure email is lowercase

		// Validate required fields
		if request.FirstName == "" {
			e["first_name"] = "First name is required"
		}
		if request.LastName == "" {
			e["last_name"] = "Last name is required"
		}
		if request.Email == "" {
			e["email"] = "Email is required"
		}
		if len(request.Email) > 255 {
			e["email"] = "Email is too long"
		}
		if request.Phone == "" {
			e["phone"] = "Phone is required"
		}
		if request.Country == "" {
			e["country"] = "Country is required"
		}
		if request.Timezone == "" {
			e["timezone"] = "Timezone is required"
		}
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		uc.logger.Warn("Failed validation for update me request", zap.Any("errors", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Call repository to update user profile in cloud
	//

	uc.logger.Debug("Executing update me in cloud use case",
		zap.String("email", request.Email),
		zap.String("first_name", request.FirstName),
		zap.String("last_name", request.LastName))

	response, err := uc.repository.UpdateMeInCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to update user profile in cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to update user profile in the cloud", err)
	}

	//
	// STEP 3: Return response
	//

	uc.logger.Info("Successfully updated user profile in cloud",
		zap.String("email", response.Email),
		zap.String("name", response.Name))

	return response, nil
}
