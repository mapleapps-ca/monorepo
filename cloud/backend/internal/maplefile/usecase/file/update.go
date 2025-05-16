// cloud/backend/internal/maplefile/usecase/file/update.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UpdateFileUseCase interface {
	Execute(ctx context.Context, file *dom_file.File) error
}

type updateFileUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewUpdateFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) UpdateFileUseCase {
	return &updateFileUseCaseImpl{config, logger, repo}
}

func (uc *updateFileUseCaseImpl) Execute(ctx context.Context, file *dom_file.File) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if file == nil {
		e["file"] = "File is required"
	} else {
		if file.ID == "" {
			e["id"] = "File ID is required"
		}
		if file.OwnerID == "" {
			e["owner_id"] = "Owner ID is required"
		}
		if file.CollectionID == "" {
			e["collection_id"] = "Collection ID is required"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file update",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Update in database.
	//

	return uc.repo.Update(file)
}
