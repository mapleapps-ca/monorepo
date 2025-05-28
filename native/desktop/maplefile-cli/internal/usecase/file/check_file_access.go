package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// CheckFileAccessUseCase defines the interface for checking user access to a file
type CheckFileAccessUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error)
}

// checkFileAccessUseCase implements the CheckFileAccessUseCase interface
type checkFileAccessUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewCheckFileAccessUseCase creates a new use case for checking file access
func NewCheckFileAccessUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) CheckFileAccessUseCase {
	logger = logger.Named("CheckFileAccessUseCase")
	return &checkFileAccessUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute checks if a user has access to a local file
func (uc *checkFileAccessUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
	userID primitive.ObjectID,
) (bool, error) {
	// Validate inputs
	if fileID.IsZero() {
		return false, errors.NewAppError("file ID is required", nil)
	}

	if userID.IsZero() {
		return false, errors.NewAppError("user ID is required", nil)
	}

	// Check if the user has access to the file
	hasAccess, err := uc.repository.CheckIfUserHasAccess(ctx, fileID, userID)
	if err != nil {
		return false, errors.NewAppError("failed to check user access to local file", err)
	}

	return hasAccess, nil
}
