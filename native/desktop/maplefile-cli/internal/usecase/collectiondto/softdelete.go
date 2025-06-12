// monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto/softdelete.go
package collectiondto

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// SoftDeleteCollectionFromCloudUseCase defines the interface for deleting a collection from cloud
type SoftDeleteCollectionFromCloudUseCase interface {
	Execute(ctx context.Context, collectionID gocql.UUID) error
}

// softDeleteCollectionFromCloudUseCase implements the SoftDeleteCollectionFromCloudUseCase interface
type softDeleteCollectionFromCloudUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewSoftDeleteCollectionFromCloudUseCase creates a new use case for deleting collections from cloud
func NewSoftDeleteCollectionFromCloudUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) SoftDeleteCollectionFromCloudUseCase {
	logger = logger.Named("SoftSoftDeleteCollectionFromCloudUseCase")
	return &softDeleteCollectionFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute deletes a collection from the cloud
func (uc *softDeleteCollectionFromCloudUseCase) Execute(ctx context.Context, collectionID gocql.UUID) error {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if collectionID.String() == "" {
		e["collection_id"] = "Collection ID is required"
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Call repository to delete collection from cloud
	//

	uc.logger.Debug("Executing (soft)delete collection from cloud use case",
		zap.String("collectionID", collectionID.String()))

	err := uc.repository.SoftDeleteInCloudByID(ctx, collectionID)
	if err != nil {
		uc.logger.Error("Failed to (soft)delete collection from cloud",
			zap.Error(err),
			zap.String("collectionID", collectionID.String()))
		return errors.NewAppError("failed to (soft)delete collection from the cloud", err)
	}

	//
	// STEP 3: Log successful deletion
	//

	uc.logger.Info("Successfully (soft)deleted collection from cloud",
		zap.String("collectionID", collectionID.String()))

	return nil
}
