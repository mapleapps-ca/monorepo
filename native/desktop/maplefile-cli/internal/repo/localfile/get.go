// native/desktop/maplefile-cli/internal/repo/localfile/get.go
package localfile

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetByID retrieves a local file by ID
func (r *localFileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*localfile.LocalFile, error) {
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
	file, err := localfile.NewFromDeserialized(fileBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize file metadata", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize file metadata", err)
	}

	r.logger.Debug("Successfully retrieved file metadata from local storage",
		zap.String("fileID", id.Hex()))
	return file, nil
}

// GetByRemoteID retrieves a local file by its remote ID
func (r *localFileRepository) GetByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*localfile.LocalFile, error) {
	r.logger.Debug("Retrieving file by remote ID from local storage", zap.String("remoteID", remoteID.Hex()))

	// Get all files and filter
	files, err := r.List(ctx, localfile.LocalFileFilter{})
	if err != nil {
		return nil, errors.NewAppError("failed to list local files", err)
	}

	// Find the file with matching remote ID
	for _, file := range files {
		if file.RemoteID == remoteID {
			return file, nil
		}
	}

	r.logger.Debug("File with remote ID not found in local storage", zap.String("remoteID", remoteID.Hex()))
	return nil, nil
}
