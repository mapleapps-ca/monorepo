// cloud/backend/internal/maplefile/repo/file/delete.go
package file

import (
	"go.uber.org/zap"
)

// Delete implements the FileRepository.Delete method
func (repo *fileRepositoryImpl) Delete(id string) error {
	// First get the file to get its storage path
	file, err := repo.metadata.Get(id)
	if err != nil {
		return err
	}
	if file == nil {
		repo.logger.Info("file not found for deletion", zap.String("id", id))
		return nil
	}

	// Delete from S3
	if err := repo.storage.DeleteEncryptedData(file.StoragePath); err != nil {
		repo.logger.Error("failed to delete file data",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	// Delete metadata
	return repo.metadata.Delete(id)
}
