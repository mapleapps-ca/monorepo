// monorepo/cloud/backend/internal/maplefile/repo/fileobjectstorage/get_object_size.go
package fileobjectstorage

import (
	"context"

	"go.uber.org/zap"
)

// GetObjectSize returns the size in bytes of an object at the given storage path
func (impl *fileObjectStorageRepositoryImpl) GetObjectSize(storagePath string) (int64, error) {
	ctx := context.Background()

	// Get object size from storage
	size, err := impl.Storage.GetObjectSize(ctx, storagePath)
	if err != nil {
		impl.Logger.Error("Failed to get object size",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return 0, err
	}

	impl.Logger.Debug("Retrieved object size",
		zap.String("storagePath", storagePath),
		zap.Int64("size", size))

	return size, nil
}
