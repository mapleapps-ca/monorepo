// internal/usecase/localfile/create_files.go
package localfile

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CreateFilesUseCase defines the interface for creating multiple local files
type CreateFilesUseCase interface {
	Execute(ctx context.Context, filesData []*file.File) ([]*file.File, error)
}

// createFilesUseCase implements the CreateFilesUseCase interface
type createFilesUseCase struct {
	logger            *zap.Logger
	fileCreateUseCase fileUseCase.CreateFilesUseCase
}

// NewCreateFilesUseCase creates a new use case for creating multiple local files
func NewCreateFilesUseCase(
	logger *zap.Logger,
	fileCreateUseCase fileUseCase.CreateFilesUseCase,
) CreateFilesUseCase {
	return &createFilesUseCase{
		logger:            logger,
		fileCreateUseCase: fileCreateUseCase,
	}
}

// Execute creates multiple local files
func (uc *createFilesUseCase) Execute(
	ctx context.Context,
	filesData []*file.File,
) ([]*file.File, error) {
	uc.logger.Debug("Creating multiple local files", zap.Int("count", len(filesData)))

	err := uc.fileCreateUseCase.Execute(ctx, filesData)
	if err != nil {
		return nil, errors.NewAppError("failed to create local files", err)
	}

	// Return the created files (with updated IDs and timestamps)
	return filesData, nil
}
