// monorepo/cloud/backend/internal/maplefile/repo/fileobjectstorage/delete.go
package fileobjectstorage

import (
	"context"

	"go.uber.org/zap"
)

// DeleteEncryptedData removes encrypted file data from S3
func (impl *fileObjectStorageRepositoryImpl) DeleteEncryptedData(storagePath string) error {
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
