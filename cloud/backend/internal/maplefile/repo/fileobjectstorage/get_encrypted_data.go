// cloud/backend/internal/maplefile/repo/fileobjectstorage/get_encrypted_data.go
package fileobjectstorage

import (
	"context"
	"io"

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
