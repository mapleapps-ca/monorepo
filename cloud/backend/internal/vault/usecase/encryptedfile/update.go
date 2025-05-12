// cloud/backend/internal/vault/usecase/encryptedfile/update.go
package encryptedfile

import (
	"context"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// UpdateEncryptedFileUseCase defines operations for updating an encrypted file
type UpdateEncryptedFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID, encryptedMetadata string, encryptedHash string, encryptedContent io.Reader) (*domain.EncryptedFile, error)
}

type updateEncryptedFileUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository domain.Repository
}

// NewUpdateEncryptedFileUseCase creates a new instance of the use case
func NewUpdateEncryptedFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repository domain.Repository,
) UpdateEncryptedFileUseCase {
	return &updateEncryptedFileUseCaseImpl{
		config:     config,
		logger:     logger.With(zap.String("component", "update-encrypted-file-usecase")),
		repository: repository,
	}
}

// Execute updates an encrypted file
func (uc *updateEncryptedFileUseCaseImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
	encryptedMetadata string,
	encryptedHash string,
	encryptedContent io.Reader,
) (*domain.EncryptedFile, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID cannot be empty")
	}

	// Get the existing file
	existingFile, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get existing file for update",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get existing file: %w", err)
	}

	if existingFile == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Update the file fields
	if encryptedMetadata != "" {
		existingFile.EncryptedMetadata = encryptedMetadata
	}

	if encryptedHash != "" {
		existingFile.EncryptedHash = encryptedHash
	}

	// Update the file
	err = uc.repository.UpdateByID(ctx, existingFile, encryptedContent)
	if err != nil {
		uc.logger.Error("Failed to update encrypted file",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update encrypted file: %w", err)
	}

	// Get the updated file
	updatedFile, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get updated file after update",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get updated file: %w", err)
	}

	uc.logger.Info("Successfully updated encrypted file",
		zap.String("id", id.Hex()),
	)

	return updatedFile, nil
}
