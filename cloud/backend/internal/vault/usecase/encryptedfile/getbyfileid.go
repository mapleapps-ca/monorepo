// cloud/backend/internal/vault/usecase/encryptedfile/getbyfileid.go
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

// GetEncryptedFileByFileIDUseCase defines operations for retrieving an encrypted file by user ID and file ID
type GetEncryptedFileByFileIDUseCase interface {
	Execute(ctx context.Context, userID primitive.ObjectID, fileID string) (*domain.EncryptedFile, error)
}

type getEncryptedFileByFileIDUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository domain.Repository
}

// NewGetEncryptedFileByFileIDUseCase creates a new instance of the use case
func NewGetEncryptedFileByFileIDUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repository domain.Repository,
) GetEncryptedFileByFileIDUseCase {
	return &getEncryptedFileByFileIDUseCaseImpl{
		config:     config,
		logger:     logger.With(zap.String("component", "get-encrypted-file-by-file-id-usecase")),
		repository: repository,
	}
}

// Execute retrieves an encrypted file by user ID and file ID
func (uc *getEncryptedFileByFileIDUseCaseImpl) Execute(
	ctx context.Context,
	userID primitive.ObjectID,
	fileID string,
) (*domain.EncryptedFile, error) {
	// Validate inputs
	if userID.IsZero() {
		return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID cannot be empty")
	}

	if fileID == "" {
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID cannot be empty")
	}

	// Retrieve the file
	file, err := uc.repository.GetByFileID(ctx, userID, fileID)
	if err != nil {
		uc.logger.Error("Failed to get encrypted file by file ID",
			zap.String("userID", userID.Hex()),
			zap.String("fileID", fileID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get encrypted file: %w", err)
	}

	if file == nil {
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File not found")
	}

	return file, nil
}
