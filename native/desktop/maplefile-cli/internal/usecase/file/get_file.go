// internal/usecase/file/get_file.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// GetFileUseCase defines the interface for getting a local file
type GetFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*file.File, error)
}

// getFileUseCase implements the GetFileUseCase interface
type getFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewGetFileUseCase creates a new use case for getting local files
func NewGetFileUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) GetFileUseCase {
	logger = logger.Named("GetFileUseCase")
	return &getFileUseCase{
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

	// Developers Note: Yes we can return `file=nil`, this is not a mistake but a deliberate decision.
	return file, nil
}
