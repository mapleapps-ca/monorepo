// internal/usecase/localfile/check_file_exists.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CheckFileExistsUseCase defines the interface for checking if a local file exists
type CheckFileExistsUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) (bool, error)
}

// checkFileExistsUseCase implements the CheckFileExistsUseCase interface
type checkFileExistsUseCase struct {
	logger             *zap.Logger
	checkExistsUseCase fileUseCase.CheckFileExistsUseCase
}

// NewCheckFileExistsUseCase creates a new use case for checking file existence
func NewCheckFileExistsUseCase(
	logger *zap.Logger,
	checkExistsUseCase fileUseCase.CheckFileExistsUseCase,
) CheckFileExistsUseCase {
	return &checkFileExistsUseCase{
		logger:             logger,
		checkExistsUseCase: checkExistsUseCase,
	}
}

// Execute checks if a file exists in the database
func (uc *checkFileExistsUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
) (bool, error) {
	uc.logger.Debug("Checking if local file exists", zap.String("fileID", fileID.Hex()))

	exists, err := uc.checkExistsUseCase.Execute(ctx, fileID)
	if err != nil {
		return false, errors.NewAppError("failed to check if local file exists", err)
	}

	return exists, nil
}
