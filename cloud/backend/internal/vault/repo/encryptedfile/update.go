// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// UpdateByID updates an encrypted file
func (repo *encryptedFileRepository) UpdateByID(
	ctx context.Context,
	file *domain.EncryptedFile,
	encryptedContent io.Reader,
) error {
	// Get the existing file to retrieve the storage path
	existingFile, err := repo.GetByID(ctx, file.ID)
	if err != nil {
		return err
	}

	if existingFile == nil {
		return fmt.Errorf("file not found")
	}

	// Update modification time
	file.ModifiedAt = time.Now()
	file.CreatedAt = existingFile.CreatedAt // Preserve creation time

	// Use the existing storage path
	file.StoragePath = existingFile.StoragePath

	// If no new content, keep the existing size
	file.EncryptedSize = existingFile.EncryptedSize

	// Update the metadata in MongoDB
	_, err = repo.collection.ReplaceOne(
		ctx,
		bson.M{"_id": file.ID},
		file,
	)

	if err != nil {
		return fmt.Errorf("failed to update encrypted file metadata: %w", err)
	}

	repo.logger.Debug("Successfully updated encrypted file",
		zap.String("id", file.ID.Hex()),
		zap.String("userID", file.UserID.Hex()),
		zap.String("fileID", file.FileID),
	)

	return nil
}
