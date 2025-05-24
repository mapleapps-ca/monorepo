// native/desktop/maplefile-cli/internal/repo/file/get.go
package file

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetByID retrieves a local file by ID
func (r *fileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*dom_file.File, error) {
	r.logger.Debug("Retrieving file from local storage", zap.String("fileID", id.Hex()))

	// Generate key for this file
	key := r.generateKey(id.Hex())

	// Get from database
	fileBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve file metadata from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve file metadata from local storage", err)
	}

	// Check if file was found
	if fileBytes == nil {
		r.logger.Warn("File metadata not found in local storage", zap.String("fileID", id.Hex()))
		return nil, nil
	}

	// Deserialize the file
	file, err := file.NewFromDeserialized(fileBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize file metadata", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize file metadata", err)
	}

	r.logger.Debug("Successfully retrieved file metadata from local storage",
		zap.String("fileID", id.Hex()))
	return file, nil
}
