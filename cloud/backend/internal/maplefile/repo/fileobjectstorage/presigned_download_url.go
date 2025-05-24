// cloud/backend/internal/maplefile/repo/fileobjectstorage/get_object_size.go
package fileobjectstorage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// GetObjectSize returns the size in bytes of an object at the given storage path
func (impl *fileObjectStorageRepositoryImpl) GeneratePresignedDownloadURL(storagePath string, duration time.Duration) (string, error) {
	ctx := context.Background()

	// Get object size from storage
	url, err := impl.Storage.GetDownloadablePresignedURL(ctx, storagePath, duration)
	if err != nil {
		impl.Logger.Error("Failed to get object size",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return "", err
	}

	return url, nil
}
