// internal/usecase/collectiondto/get.go
package collectiondto

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// GetCollectionFromCloudUseCase defines the interface for creating a cloud collection
type GetCollectionFromCloudUseCase interface {
	Execute(ctx context.Context, collectionID gocql.UUID) (*collectiondto.CollectionDTO, error)
}

// getCollectionFromCloudUseCase implements the GetCollectionFromCloudUseCase interface
type getCollectionFromCloudUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewGetCollectionFromCloudUseCase creates a new use case for creating cloud collections
func NewGetCollectionFromCloudUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) GetCollectionFromCloudUseCase {
	logger = logger.Named("GetCollectionFromCloudUseCase")
	return &getCollectionFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new cloud collection
func (uc *getCollectionFromCloudUseCase) Execute(ctx context.Context, collectionID gocql.UUID) (*collectiondto.CollectionDTO, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if collectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Submit our collection to the cloud.
	//

	// Call the repository to get the collection
	cloudCollectionDTO, err := uc.repository.GetFromCloudByID(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection from the cloud", err)
	}

	//
	// STEP 3: Return our collection response from the cloud.
	//

	return cloudCollectionDTO, nil
}
