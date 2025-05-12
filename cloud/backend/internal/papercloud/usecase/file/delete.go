// cloud/backend/internal/papercloud/usecase/file/delete.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteFileUseCase interface {
	Execute(ctx context.Context, id string) error
}

type deleteFileUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewDeleteFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) DeleteFileUseCase {
	return &deleteFileUseCaseImpl{config, logger, repo}
}

func (uc *deleteFileUseCaseImpl) Execute(ctx context.Context, id string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id == "" {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file deletion",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Delete from database and storage.
	//

	// Repository will handle deleting both the metadata and the actual file content
	return uc.repo.Delete(id)
}
