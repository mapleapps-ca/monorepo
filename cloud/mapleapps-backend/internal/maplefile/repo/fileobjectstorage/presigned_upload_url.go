// monorepo/cloud/backend/internal/maplefile/repo/fileobjectstorage/presigned_upload_url.go
package fileobjectstorage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// GeneratePresignedUploadURL creates a temporary, time-limited URL that allows clients to upload
// encrypted file data directly to the storage system at the specified storage path.
func (impl *fileObjectStorageRepositoryImpl) GeneratePresignedUploadURL(storagePath string, duration time.Duration) (string, error) {
	ctx := context.Background()

	// Generate presigned upload URL
	url, err := impl.Storage.GeneratePresignedUploadURL(ctx, storagePath, duration)
	if err != nil {
		impl.Logger.Error("Failed to generate presigned upload URL",
			zap.String("storagePath", storagePath),
			zap.Duration("duration", duration),
			zap.Error(err))
		return "", err
	}

	impl.Logger.Debug("Generated presigned upload URL",
		zap.String("storagePath", storagePath),
		zap.Duration("duration", duration))

	return url, nil
}
