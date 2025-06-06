// cloud/backend/internal/maplefile/usecase/filemetadata/delete_many.go
package filemetadata

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type DeleteManyFileMetadataUseCase interface {
	Execute(ids []primitive.ObjectID) error
}

type deleteManyFileMetadataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewDeleteManyFileMetadataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) DeleteManyFileMetadataUseCase {
	logger = logger.Named("DeleteManyFileMetadataUseCase")
	return &deleteManyFileMetadataUseCaseImpl{config, logger, repo}
}

func (uc *deleteManyFileMetadataUseCaseImpl) Execute(ids []primitive.ObjectID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ids == nil || len(ids) == 0 {
		e["ids"] = "File IDs are required"
	} else {
		for i, id := range ids {
			if id.IsZero() {
				e[fmt.Sprintf("ids[%d]", i)] = "File ID is required"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata batch deletion",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Delete from database.
	//

	return uc.repo.SoftDeleteMany(ids)
}
