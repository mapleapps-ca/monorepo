// cloud/backend/internal/maplefile/usecase/filemetadata/get_by_collection.go
package filemetadata

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileMetadataByCollectionUseCase interface {
	Execute(collectionID primitive.ObjectID) ([]*dom_file.File, error)
}

type getFileMetadataByCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataByCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataByCollectionUseCase {
	logger = logger.Named("GetFileMetadataByCollectionUseCase")
	return &getFileMetadataByCollectionUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataByCollectionUseCaseImpl) Execute(collectionID primitive.ObjectID) ([]*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval by collection",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByCollection(collectionID)
}
