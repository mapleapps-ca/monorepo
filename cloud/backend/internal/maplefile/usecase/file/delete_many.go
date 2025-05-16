// cloud/backend/internal/maplefile/usecase/file/delete_many.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteManyFilesUseCase interface {
	Execute(ctx context.Context, ids []primitive.ObjectID) error
}

type deleteManyFilesUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewDeleteManyFilesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) DeleteManyFilesUseCase {
	return &deleteManyFilesUseCaseImpl{config, logger, repo}
}

func (uc *deleteManyFilesUseCaseImpl) Execute(ctx context.Context, ids []primitive.ObjectID) error {
	//
	// STEP 1: Validation.
	//

	if len(ids) == 0 {
		return nil // Nothing to delete
	}

	// Check if any of the IDs are zero
	for i, id := range ids {
		if id.IsZero() {
			uc.logger.Warn("Invalid file ID in batch deletion",
				zap.Int("index", i))
			return httperror.NewForBadRequestWithSingleField("ids", "All file IDs must be valid")
		}
	}

	//
	// STEP 2: Delete from database and storage.
	//

	// Repository will handle deleting both the metadata and the actual file content for all files
	return uc.repo.DeleteMany(ids)
}
