// internal/usecase/file/get_files_by_ids.go
package file

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// GetFilesByIDsUseCase defines the interface for getting multiple local files
type GetFilesByIDsUseCase interface {
	Execute(ctx context.Context, ids []primitive.ObjectID) ([]*dom_file.File, error)
}

// getFilesByIDsUseCase implements the GetFilesByIDsUseCase interface
type getFilesByIDsUseCase struct {
	logger     *zap.Logger
	repository dom_file.FileRepository
}

// NewGetFilesByIDsUseCase creates a new use case for getting multiple local files
func NewGetFilesByIDsUseCase(
	logger *zap.Logger,
	repository dom_file.FileRepository,
) GetFilesByIDsUseCase {
	return &getFilesByIDsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves multiple local files by IDs
func (uc *getFilesByIDsUseCase) Execute(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]*dom_file.File, error) {
	if len(ids) == 0 {
		return []*dom_file.File{}, nil
	}

	// Validate all IDs
	for i, id := range ids {
		if id.IsZero() {
			return nil, errors.NewAppError(fmt.Sprintf("file ID at index %d is invalid", i), nil)
		}
	}

	// Get the files from the repository
	files, err := uc.repository.GetByIDs(ctx, ids)
	if err != nil {
		return nil, errors.NewAppError("failed to get local files", err)
	}

	return files, nil
}
