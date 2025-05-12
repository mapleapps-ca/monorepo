// cloud/backend/internal/papercloud/repo/file/storage/delete.go
package storage

import (
	"context"

	"go.uber.org/zap"
)

// DeleteEncryptedData removes encrypted file data from S3
func (impl *fileStorageRepositoryImpl) DeleteEncryptedData(storagePath string) error {
	ctx := context.Background()

	// Delete the encrypted data
	err := impl.Storage.DeleteByKeys(ctx, []string{storagePath})
	if err != nil {
		impl.Logger.Error("Failed to delete encrypted data",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return err
	}

	return nil
}
