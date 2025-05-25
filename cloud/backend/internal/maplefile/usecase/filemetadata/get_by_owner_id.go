// cloud/backend/internal/maplefile/usecase/filemetadata/get_by_owner_id.go
package filemetadata

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileMetadataByOwnerIDUseCase interface {
	Execute(ownerID primitive.ObjectID) ([]*dom_file.File, error)
}

type getFileMetadataByOwnerIDUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataByOwnerIDUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataByOwnerIDUseCase {
	return &getFileMetadataByOwnerIDUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataByOwnerIDUseCaseImpl) Execute(ownerID primitive.ObjectID) ([]*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ownerID.IsZero() {
		e["owner_id"] = "Created by user ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval by owner_id",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByOwnerID(ownerID)
}
