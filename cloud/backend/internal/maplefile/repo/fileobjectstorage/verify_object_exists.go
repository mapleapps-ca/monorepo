// cloud/backend/internal/maplefile/repo/fileobjectstorage/get_object_size.go
package fileobjectstorage

import (
	"context"

	"go.uber.org/zap"
)

func (impl *fileObjectStorageRepositoryImpl) VerifyObjectExists(storagePath string) (bool, error) {
	ctx := context.Background()

	// Get object size from storage
	exists, err := impl.Storage.ObjectExists(ctx, storagePath)
	if err != nil {
		impl.Logger.Error("Failed to verify if object exists",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return false, err
	}

	return exists, nil
}
