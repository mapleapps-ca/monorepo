// cloud/backend/internal/maplefile/repo/fileobjectstorage/download.go
package fileobjectstorage

import (
	"context"
	"io"
	"time"

	"go.uber.org/zap"
)

// GetEncryptedData retrieves encrypted file data from S3
func (impl *fileObjectStorageRepositoryImpl) GetEncryptedData(storagePath string) ([]byte, error) {
	ctx := context.Background()

	// Get the encrypted data
	reader, err := impl.Storage.GetBinaryData(ctx, storagePath)
	if err != nil {
		impl.Logger.Error("Failed to get encrypted data",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return nil, err
	}
	defer reader.Close()

	// Read all data into memory
	data, err := io.ReadAll(reader)
	if err != nil {
		impl.Logger.Error("Failed to read encrypted data",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return nil, err
	}

	return data, nil
}

// GeneratePresignedURL creates a time-limited URL for downloading the file directly
func (impl *fileObjectStorageRepositoryImpl) GeneratePresignedURL(storagePath string, duration time.Duration) (string, error) {
	ctx := context.Background()

	// Generate presigned URL
	url, err := impl.Storage.GetPresignedURL(ctx, storagePath, duration)
	if err != nil {
		impl.Logger.Error("Failed to generate presigned URL",
			zap.String("storagePath", storagePath),
			zap.Error(err))
		return "", err
	}

	return url, nil
}
