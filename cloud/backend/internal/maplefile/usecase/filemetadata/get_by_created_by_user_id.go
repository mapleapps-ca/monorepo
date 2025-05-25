// cloud/backend/internal/maplefile/usecase/filemetadata/get_by_created_by_user_id.go
package filemetadata

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileMetadataByCreatedByUserIDUseCase interface {
	Execute(createdByUserID primitive.ObjectID) ([]*dom_file.File, error)
}

type getFileMetadataByCreatedByUserIDUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataByCreatedByUserIDUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataByCreatedByUserIDUseCase {
	return &getFileMetadataByCreatedByUserIDUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataByCreatedByUserIDUseCaseImpl) Execute(createdByUserID primitive.ObjectID) ([]*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if createdByUserID.IsZero() {
		e["created_by_user_id"] = "Created by user ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval by created_by_user_id",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByCreatedByUserID(createdByUserID)
}
