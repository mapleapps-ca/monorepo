// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// Create stores a new encrypted file
func (repo *encryptedFileRepository) Create(
	ctx context.Context,
	file *domain.EncryptedFile,
	encryptedContent io.Reader,
) error {
	// Generate a new ID if not provided
	if file.ID == primitive.NilObjectID {
		file.ID = primitive.NewObjectID()
	}

	// Set creation and modification times
	now := time.Now()
	file.CreatedAt = now
	file.ModifiedAt = now

	// Generate a unique storage path for the file
	userID := file.UserID.Hex()
	storagePath := fmt.Sprintf("%s/%s", userID, file.FileID)
	file.StoragePath = storagePath

	// Save metadata to MongoDB collection
	_, err := repo.collection.InsertOne(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to save encrypted file metadata: %w", err)
	}

	repo.logger.Debug("Successfully created encrypted file",
		zap.String("id", file.ID.Hex()),
		zap.String("userID", userID),
		zap.String("fileID", file.FileID),
	)

	return nil
}
