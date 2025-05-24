// internal/usecase/localfile/get_files_by_ids.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// GetFilesByIDsUseCase defines the interface for getting multiple local files by IDs
type GetFilesByIDsUseCase interface {
	Execute(ctx context.Context, fileIDs []primitive.ObjectID) ([]*file.File, error)
}

// getFilesByIDsUseCase implements the GetFilesByIDsUseCase interface
type getFilesByIDsUseCase struct {
	logger              *zap.Logger
	fileGetByIDsUseCase fileUseCase.GetFilesByIDsUseCase
}

// NewGetFilesByIDsUseCase creates a new use case for getting files by IDs
func NewGetFilesByIDsUseCase(
	logger *zap.Logger,
	fileGetByIDsUseCase fileUseCase.GetFilesByIDsUseCase,
) GetFilesByIDsUseCase {
	return &getFilesByIDsUseCase{
		logger:              logger,
		fileGetByIDsUseCase: fileGetByIDsUseCase,
	}
}

// Execute retrieves multiple files by their IDs
func (uc *getFilesByIDsUseCase) Execute(
	ctx context.Context,
	fileIDs []primitive.ObjectID,
) ([]*file.File, error) {
	uc.logger.Debug("Getting local files by IDs", zap.Int("count", len(fileIDs)))

	files, err := uc.fileGetByIDsUseCase.Execute(ctx, fileIDs)
	if err != nil {
		return nil, errors.NewAppError("failed to get local files by IDs", err)
	}

	return files, nil
}
