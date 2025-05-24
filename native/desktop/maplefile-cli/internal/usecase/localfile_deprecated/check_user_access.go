// internal/usecase/localfile/check_user_access.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CheckUserAccessUseCase defines the interface for checking user access to a local file
type CheckUserAccessUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error)
}

// checkUserAccessUseCase implements the CheckUserAccessUseCase interface
type checkUserAccessUseCase struct {
	logger             *zap.Logger
	checkAccessUseCase fileUseCase.CheckFileAccessUseCase
}

// NewCheckUserAccessUseCase creates a new use case for checking user access
func NewCheckUserAccessUseCase(
	logger *zap.Logger,
	checkAccessUseCase fileUseCase.CheckFileAccessUseCase,
) CheckUserAccessUseCase {
	return &checkUserAccessUseCase{
		logger:             logger,
		checkAccessUseCase: checkAccessUseCase,
	}
}

// Execute checks if a user has access to a file
func (uc *checkUserAccessUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
	userID primitive.ObjectID,
) (bool, error) {
	uc.logger.Debug("Checking user access to local file",
		zap.String("fileID", fileID.Hex()),
		zap.String("userID", userID.Hex()))

	hasAccess, err := uc.checkAccessUseCase.Execute(ctx, fileID, userID)
	if err != nil {
		return false, errors.NewAppError("failed to check user access to local file", err)
	}

	return hasAccess, nil
}
