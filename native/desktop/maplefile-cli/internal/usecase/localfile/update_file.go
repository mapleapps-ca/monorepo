// internal/usecase/localfile/update_file.go
package localfile

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// UpdateFileUseCase defines the interface for updating a local file
type UpdateFileUseCase interface {
	Execute(ctx context.Context, input fileUseCase.UpdateFileInput) (*file.File, error)
}

// updateFileUseCase implements the UpdateFileUseCase interface
type updateFileUseCase struct {
	logger            *zap.Logger
	fileUpdateUseCase fileUseCase.UpdateFileUseCase
}

// NewUpdateFileUseCase creates a new use case for updating a local file
func NewUpdateFileUseCase(
	logger *zap.Logger,
	fileUpdateUseCase fileUseCase.UpdateFileUseCase,
) UpdateFileUseCase {
	return &updateFileUseCase{
		logger:            logger,
		fileUpdateUseCase: fileUpdateUseCase,
	}
}

// Execute updates an existing local file
func (uc *updateFileUseCase) Execute(
	ctx context.Context,
	input fileUseCase.UpdateFileInput,
) (*file.File, error) {
	uc.logger.Debug("Updating local file", zap.String("fileID", input.ID.Hex()))

	updatedFile, err := uc.fileUpdateUseCase.Execute(ctx, input)
	if err != nil {
		return nil, errors.NewAppError("failed to update local file", err)
	}

	return updatedFile, nil
}
