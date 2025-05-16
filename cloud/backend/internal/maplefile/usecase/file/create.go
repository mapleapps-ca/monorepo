// cloud/backend/internal/maplefile/usecase/file/create.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CreateFileUseCase interface {
	Execute(ctx context.Context, file *dom_file.File) error
}

type createFileUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewCreateFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) CreateFileUseCase {
	return &createFileUseCaseImpl{config, logger, repo}
}

func (uc *createFileUseCaseImpl) Execute(ctx context.Context, file *dom_file.File) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if file == nil {
		e["file"] = "File is required"
	} else {
		if file.OwnerID.IsZero() {
			e["owner_id"] = "Owner ID is required"
		}
		if file.CollectionID.IsZero() {
			e["collection_id"] = "Collection ID is required"
		}
		// EncryptedFileID is optional as it may be generated in the repository if not provided
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file creation",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Insert into database.
	//

	return uc.repo.Create(file)
}
