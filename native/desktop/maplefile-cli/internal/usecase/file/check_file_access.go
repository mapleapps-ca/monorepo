package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// CheckFileAccessUseCase defines the interface for checking user access to a file
type CheckFileAccessUseCase interface {
	Execute(ctx context.Context, fileID gocql.UUID, userID gocql.UUID) (bool, error)
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
	fileID gocql.UUID,
	userID gocql.UUID,
) (bool, error) {
	// Validate inputs
	if fileID.String() == "" {
		return false, errors.NewAppError("file ID is required", nil)
	}

	if userID.String() == "" {
		return false, errors.NewAppError("user ID is required", nil)
	}

	// Check if the user has access to the file
	hasAccess, err := uc.repository.CheckIfUserHasAccess(ctx, fileID, userID)
	if err != nil {
		return false, errors.NewAppError("failed to check user access to local file", err)
	}

	return hasAccess, nil
}
