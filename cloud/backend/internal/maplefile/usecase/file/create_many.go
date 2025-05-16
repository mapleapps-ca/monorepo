// cloud/backend/internal/maplefile/usecase/file/create_many.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CreateManyFilesUseCase interface {
	Execute(ctx context.Context, files []*dom_file.File) error
}

type createManyFilesUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewCreateManyFilesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) CreateManyFilesUseCase {
	return &createManyFilesUseCaseImpl{config, logger, repo}
}

func (uc *createManyFilesUseCaseImpl) Execute(ctx context.Context, files []*dom_file.File) error {
	//
	// STEP 1: Validation.
	//

	if len(files) == 0 {
		return nil // Nothing to create
	}

	var validationErrors []map[string]string
	isValid := true

	for i, file := range files {
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

		if len(e) > 0 {
			e["index"] = string(i)
			validationErrors = append(validationErrors, e)
			isValid = false
		}
	}

	if !isValid {
		uc.logger.Warn("Failed validating batch file creation",
			zap.Any("errors", validationErrors))
		return httperror.NewForBadRequestWithSingleField("files", "Invalid file data in batch")
	}

	//
	// STEP 2: Insert into database.
	//

	return uc.repo.CreateMany(files)
}
