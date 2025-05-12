// cloud/backend/internal/vault/usecase/encryptedfile/list.go
package encryptedfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// ListEncryptedFilesUseCase defines operations for listing all encrypted files for a user
type ListEncryptedFilesUseCase interface {
	Execute(ctx context.Context, userID primitive.ObjectID) ([]*domain.EncryptedFile, error)
}

type listEncryptedFilesUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository domain.Repository
}

// NewListEncryptedFilesUseCase creates a new instance of the use case
func NewListEncryptedFilesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repository domain.Repository,
) ListEncryptedFilesUseCase {
	return &listEncryptedFilesUseCaseImpl{
		config:     config,
		logger:     logger.With(zap.String("component", "list-encrypted-files-usecase")),
		repository: repository,
	}
}

// Execute lists all encrypted files for a user
func (uc *listEncryptedFilesUseCaseImpl) Execute(
	ctx context.Context,
	userID primitive.ObjectID,
) ([]*domain.EncryptedFile, error) {
	// Validate inputs
	if userID.IsZero() {
		return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID cannot be empty")
	}

	// List the files
	files, err := uc.repository.ListByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to list encrypted files",
			zap.String("userID", userID.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list encrypted files: %w", err)
	}

	uc.logger.Debug("Successfully listed encrypted files",
		zap.String("userID", userID.Hex()),
		zap.Int("count", len(files)),
	)

	return files, nil
}
