// cloud/backend/internal/maplefile/repo/file/delete.go
package file

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Delete implements the FileRepository.Delete method
func (repo *fileRepositoryImpl) Delete(id primitive.ObjectID) error {
	// First get the file to get its storage path
	file, err := repo.metadata.Get(id)
	if err != nil {
		return err
	}
	if file == nil {
		repo.logger.Info("file not found for deletion", zap.Any("id", id))
		return nil
	}

	// Delete from S3 if storage path exists
	if file.FileObjectKey != "" {
		if err := repo.storage.DeleteEncryptedData(file.FileObjectKey); err != nil {
			repo.logger.Error("failed to delete file data",
				zap.Any("id", id),
				zap.Error(err))
			return err
		}
	}

	// Delete thumbnail if it exists
	if file.ThumbnailObjectKey != "" {
		if err := repo.storage.DeleteEncryptedData(file.ThumbnailObjectKey); err != nil {
			repo.logger.Warn("failed to delete thumbnail",
				zap.Any("id", id),
				zap.Error(err))
			// Continue with deletion even if thumbnail deletion fails
		}
	}

	// Delete metadata
	return repo.metadata.Delete(id)
}

// DeleteMany implements the FileRepository.DeleteMany method
func (repo *fileRepositoryImpl) DeleteMany(ids []primitive.ObjectID) error {
	// Get all files first to get their storage paths
	files, err := repo.metadata.GetByIDs(ids)
	if err != nil {
		return err
	}

	// Delete all files from S3
	for _, file := range files {
		// Delete main file data
		if file.FileObjectKey != "" {
			if err := repo.storage.DeleteEncryptedData(file.FileObjectKey); err != nil {
				repo.logger.Error("failed to delete file data during batch deletion",
					zap.Any("id", file.ID),
					zap.Error(err))
				// Continue with other deletions
			}
		}

		// Delete thumbnail if it exists
		if file.ThumbnailObjectKey != "" {
			if err := repo.storage.DeleteEncryptedData(file.ThumbnailObjectKey); err != nil {
				repo.logger.Warn("failed to delete thumbnail during batch deletion",
					zap.Any("id", file.ID),
					zap.Error(err))
				// Continue with deletion even if thumbnail deletion fails
			}
		}
	}

	// Delete all metadata
	return repo.metadata.DeleteMany(ids)
}
