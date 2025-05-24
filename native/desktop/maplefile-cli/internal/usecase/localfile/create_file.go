// internal/usecase/localfile/create_file.go
package localfile

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CreateFileUseCase defines the interface for creating a local file
type CreateFileUseCase interface {
	Execute(ctx context.Context, fileData *file.File) (*file.File, error)
}

// createFileUseCase implements the CreateFileUseCase interface
type createFileUseCase struct {
	logger            *zap.Logger
	fileCreateUseCase fileUseCase.CreateFileUseCase
}

// NewCreateFileUseCase creates a new use case for creating a local file
func NewCreateFileUseCase(
	logger *zap.Logger,
	fileCreateUseCase fileUseCase.CreateFileUseCase,
) CreateFileUseCase {
	return &createFileUseCase{
		logger:            logger,
		fileCreateUseCase: fileCreateUseCase,
	}
}

// Execute creates a new local file
func (uc *createFileUseCase) Execute(
	ctx context.Context,
	fileData *file.File,
) (*file.File, error) {
	uc.logger.Debug("Creating local file", zap.String("fileName", fileData.Name))

	err := uc.fileCreateUseCase.Execute(ctx, fileData)
	if err != nil {
		return nil, errors.NewAppError("failed to create local file", err)
	}

	// Return the created file (with updated ID and timestamps)
	return fileData, nil
}
