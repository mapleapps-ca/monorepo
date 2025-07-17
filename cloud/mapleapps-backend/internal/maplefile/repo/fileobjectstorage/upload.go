// monorepo/cloud/backend/internal/maplefile/repo/fileobjectstorage/upload.go
package fileobjectstorage

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// StoreEncryptedData uploads encrypted file data to S3 and returns the storage path
func (impl *fileObjectStorageRepositoryImpl) StoreEncryptedData(ownerID string, fileID string, encryptedData []byte) (string, error) {
	ctx := context.Background()

	// Generate a storage path using a deterministic pattern
	storagePath := fmt.Sprintf("users/%s/files/%s", ownerID, fileID)

	// Always store encrypted data as private
	err := impl.Storage.UploadContentWithVisibility(ctx, storagePath, encryptedData, false)
	if err != nil {
		impl.Logger.Error("Failed to store encrypted data",
			zap.String("fileID", fileID),
			zap.String("ownerID", ownerID),
			zap.Error(err))
		return "", err
	}

	return storagePath, nil
}
