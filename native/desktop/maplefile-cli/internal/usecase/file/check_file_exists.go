// internal/usecase/file/check_file_exists.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// CheckFileExistsUseCase defines the interface for checking if a file exists
type CheckFileExistsUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (bool, error)
}

// checkFileExistsUseCase implements the CheckFileExistsUseCase interface
type checkFileExistsUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewCheckFileExistsUseCase creates a new use case for checking file existence
func NewCheckFileExistsUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) CheckFileExistsUseCase {
	return &checkFileExistsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute checks if a local file exists by ID
func (uc *checkFileExistsUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (bool, error) {
	// Validate inputs
	if id.IsZero() {
		return false, errors.NewAppError("file ID is required", nil)
	}

	// Check if the file exists
	exists, err := uc.repository.CheckIfExistsByID(ctx, id)
	if err != nil {
		return false, errors.NewAppError("failed to check if local file exists", err)
	}

	return exists, nil
}
