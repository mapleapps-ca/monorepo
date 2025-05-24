// internal/usecase/localfile/get_file.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// GetFileUseCase defines the interface for getting a single local file
type GetFileUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) (*file.File, error)
}

// getFileUseCase implements the GetFileUseCase interface
type getFileUseCase struct {
	logger         *zap.Logger
	fileGetUseCase fileUseCase.GetFileUseCase
}

// NewGetFileUseCase creates a new use case for getting a local file
func NewGetFileUseCase(
	logger *zap.Logger,
	fileGetUseCase fileUseCase.GetFileUseCase,
) GetFileUseCase {
	return &getFileUseCase{
		logger:         logger,
		fileGetUseCase: fileGetUseCase,
	}
}

// Execute retrieves a single file by ID
func (uc *getFileUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
) (*file.File, error) {
	uc.logger.Debug("Getting local file", zap.String("fileID", fileID.Hex()))

	file, err := uc.fileGetUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	return file, nil
}
