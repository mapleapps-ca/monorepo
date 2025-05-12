// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// DeleteByID deletes an encrypted file
func (repo *encryptedFileRepository) DeleteByID(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// First get the file to retrieve the storage path
	file, err := repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if file == nil {
		return fmt.Errorf("file not found")
	}

	// Delete from MongoDB collection
	_, err = repo.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete encrypted file metadata: %w", err)
	}

	repo.logger.Debug("Successfully deleted encrypted file",
		zap.String("id", id.Hex()),
		zap.String("userID", file.UserID.Hex()),
		zap.String("fileID", file.FileID),
	)

	return nil
}
