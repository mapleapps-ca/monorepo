// internal/usecase/file/get.go
package file

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// GetFileUseCase defines the interface for getting a local file
type GetFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*file.File, error)
}

// GetFilesUseCase defines the interface for getting multiple local files
type GetFilesUseCase interface {
	Execute(ctx context.Context, ids []primitive.ObjectID) ([]*file.File, error)
}

// getFileUseCase implements the GetFileUseCase interface
type getFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// getFilesUseCase implements the GetFilesUseCase interface
type getFilesUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewGetFileUseCase creates a new use case for getting local files
func NewGetFileUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) GetFileUseCase {
	return &getFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// NewGetFilesUseCase creates a new use case for getting multiple local files
func NewGetFilesUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) GetFilesUseCase {
	return &getFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves a local file by ID
func (uc *getFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*file.File, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Get the file from the repository
	file, err := uc.repository.Get(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	if file == nil {
		return nil, errors.NewAppError("local file not found", nil)
	}

	return file, nil
}

// Execute retrieves multiple local files by IDs
func (uc *getFilesUseCase) Execute(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]*file.File, error) {
	if len(ids) == 0 {
		return []*file.File{}, nil
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
