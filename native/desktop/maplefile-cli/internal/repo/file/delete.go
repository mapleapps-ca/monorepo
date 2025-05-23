// monorepo/native/desktop/maplefile-cli/internal/repo/file/delete.go
package file

import (
	"context"
	"os"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Delete removes a local file and its data
func (r *fileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	r.logger.Debug("Deleting file from local storage", zap.String("fileID", id.Hex()))

	// Get the file first to access its file paths
	file, err := r.GetByID(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to get file before deletion", err)
	}

	if file == nil {
		r.logger.Warn("File not found for deletion", zap.String("fileID", id.Hex()))
		return nil // Nothing to delete, return success
	}

	// Delete the encrypted file data from the filesystem if it exists
	if file.EncryptedFilePath != "" {
		if err := os.Remove(file.EncryptedFilePath); err != nil && !os.IsNotExist(err) {
			r.logger.Error("Failed to delete encrypted file data",
				zap.String("encryptedFilePath", file.EncryptedFilePath),
				zap.Error(err))
			// Continue with other deletions even if this one fails
		} else {
			r.logger.Debug("Deleted encrypted file data",
				zap.String("encryptedFilePath", file.EncryptedFilePath))
		}
	}

	// Delete the decrypted file data from the filesystem if it exists
	if file.DecryptedFilePath != "" {
		if err := os.Remove(file.DecryptedFilePath); err != nil && !os.IsNotExist(err) {
			r.logger.Error("Failed to delete decrypted file data",
				zap.String("decryptedFilePath", file.DecryptedFilePath),
				zap.Error(err))
			// Continue with metadata deletion even if file deletion fails
		} else {
			r.logger.Debug("Deleted decrypted file data",
				zap.String("decryptedFilePath", file.DecryptedFilePath))
		}
	}

	//TODO: REPAIR
	// // Delete thumbnail if it exists
	// if file.LocalThumbnailPath != "" {
	// 	if err := os.Remove(file.LocalThumbnailPath); err != nil && !os.IsNotExist(err) {
	// 		r.logger.Error("Failed to delete thumbnail data",
	// 			zap.String("localThumbnailPath", file.LocalThumbnailPath),
	// 			zap.Error(err))
	// 		// Continue with metadata deletion even if thumbnail deletion fails
	// 	} else {
	// 		r.logger.Debug("Deleted thumbnail data",
	// 			zap.String("localThumbnailPath", file.LocalThumbnailPath))
	// 	}
	// }

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
		zap.String("fileID", id.Hex()),
		zap.String("storageMode", file.StorageMode))
	return nil
}
