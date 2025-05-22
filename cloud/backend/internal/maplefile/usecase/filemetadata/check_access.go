// cloud/backend/internal/maplefile/usecase/filemetadata/check_access.go
package filemetadata

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CheckFileAccessUseCase interface {
	Execute(fileID, userID primitive.ObjectID) (bool, error)
}

type checkFileAccessUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewCheckFileAccessUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) CheckFileAccessUseCase {
	return &checkFileAccessUseCaseImpl{config, logger, repo}
}

func (uc *checkFileAccessUseCaseImpl) Execute(fileID, userID primitive.ObjectID) (bool, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if fileID.IsZero() {
		e["file_id"] = "File ID is required"
	}
	if userID.IsZero() {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file access check",
			zap.Any("error", e))
		return false, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Check access in database.
	//

	return uc.repo.CheckIfUserHasAccess(fileID, userID)
}
