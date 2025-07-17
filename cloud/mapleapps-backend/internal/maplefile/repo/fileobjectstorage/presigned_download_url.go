// monorepo/cloud/backend/internal/maplefile/repo/fileobjectstorage/presigned_download_url.go
package fileobjectstorage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// GeneratePresignedDownloadURL creates a time-limited URL that allows direct download
// of the file data located at the given storage path, with proper content disposition headers.
func (impl *fileObjectStorageRepositoryImpl) GeneratePresignedDownloadURL(storagePath string, duration time.Duration) (string, error) {
	ctx := context.Background()

	// Generate presigned download URL with content disposition
	url, err := impl.Storage.GetDownloadablePresignedURL(ctx, storagePath, duration)
	if err != nil {
		impl.Logger.Error("Failed to generate presigned download URL",
			zap.String("storagePath", storagePath),
			zap.Duration("duration", duration),
			zap.Error(err))
		return "", err
	}

	impl.Logger.Debug("Generated presigned download URL",
		zap.String("storagePath", storagePath),
		zap.Duration("duration", duration))

	return url, nil
}
