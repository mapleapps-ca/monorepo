// cloud/backend/internal/maplefile/repo/fileobjectstorage/verify_object_exists.go
package fileobjectstorage

import (
	"context"

	"go.uber.org/zap"
)

// VerifyObjectExists checks if an object exists at the given storage path.
func (impl *fileObjectStorageRepositoryImpl) VerifyObjectExists(storagePath string) (bool, error) {
	ctx := context.Background()

	// Check if object exists in storage
	exists, err := impl.Storage.ObjectExists(ctx, storagePath)
	if err != nil {
		impl.Logger.Error("Failed to verify if object exists",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return false, err
	}

	impl.Logger.Debug("Verified object existence",
		zap.String("storagePath", storagePath),
		zap.Bool("exists", exists))

	return exists, nil
}
