// internal/usecase/publiclookupdto/get.go
package publiclookupdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// GetPublicLookupFromCloudUseCase defines the interface for creating a cloud collection
type GetPublicLookupFromCloudUseCase interface {
	Execute(ctx context.Context, req *publiclookupdto.PublicLookupRequestDTO) (*publiclookupdto.PublicLookupResponseDTO, error)
}

// getPublicLookupFromCloudUseCase implements the GetPublicLookupFromCloudUseCase interface
type getPublicLookupFromCloudUseCase struct {
	logger     *zap.Logger
	repository publiclookupdto.PublicLookupDTORepository
}

// NewGetPublicLookupFromCloudUseCase creates a new use case for creating cloud collections
func NewGetPublicLookupFromCloudUseCase(
	logger *zap.Logger,
	repository publiclookupdto.PublicLookupDTORepository,
) GetPublicLookupFromCloudUseCase {
	logger = logger.Named("GetPublicLookupFromCloudUseCase")
	return &getPublicLookupFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new cloud collection
func (uc *getPublicLookupFromCloudUseCase) Execute(ctx context.Context, req *publiclookupdto.PublicLookupRequestDTO) (*publiclookupdto.PublicLookupResponseDTO, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if req == nil {
		return nil, httperror.NewForBadRequestWithSingleField("request", "no request data submitted")
	} else {
		if req.Email == "" {
			e["collection_id"] = "PublicLookup ID is required"
		}
	}
	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Submit our collection to the cloud.
	//

	// Call the repository to get the collection
	cloudPublicLookupDTO, err := uc.repository.GetFromCloud(ctx, req)
	if err != nil {
		return nil, errors.NewAppError("failed to execute public user lookup from the cloud", err)
	}

	//
	// STEP 3: Return our lookup response from the cloud.
	//

	return cloudPublicLookupDTO, nil
}
