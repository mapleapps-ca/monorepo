// monorepo/native/desktop/maplefile-cli/internal/repo/localfile/delete.go
package localfile

import (
	"context"
	"os"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Delete removes a local file and its data
func (r *localFileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	r.logger.Debug("Deleting file from local storage", zap.String("fileID", id.Hex()))

	// Get the file first to access its local file path
	file, err := r.GetByID(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to get file before deletion", err)
	}

	if file == nil {
		r.logger.Warn("File not found for deletion", zap.String("fileID", id.Hex()))
		return nil // Nothing to delete, return success
	}

	// Delete the actual file data from the filesystem if it exists
	if file.LocalFilePath != "" {
		if err := os.Remove(file.LocalFilePath); err != nil && !os.IsNotExist(err) {
			r.logger.Error("Failed to delete file data",
				zap.String("localFilePath", file.LocalFilePath),
				zap.Error(err))
			// Continue with metadata deletion even if file deletion fails
		}
	}

	// Delete thumbnail if it exists
	if file.LocalThumbnailPath != "" {
		if err := os.Remove(file.LocalThumbnailPath); err != nil && !os.IsNotExist(err) {
			r.logger.Error("Failed to delete thumbnail data",
				zap.String("localThumbnailPath", file.LocalThumbnailPath),
				zap.Error(err))
			// Continue with metadata deletion even if thumbnail deletion fails
		}
	}

	// Generate key for this file
	key := r.generateKey(id.Hex())

	// Delete metadata from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete file metadata from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete file metadata from local storage", err)
	}

	r.logger.Info("File deleted successfully from local storage",
		zap.String("fileID", id.Hex()))
	return nil
}
