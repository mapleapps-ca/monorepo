// native/desktop/maplefile-cli/internal/repo/file/get.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) Get(ctx context.Context, id primitive.ObjectID) (*dom_file.File, error) {
	// Generate key for this file
	key := r.generateKey(id.Hex())

	// Get from database
	fileBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve file from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve file from local storage", err)
	}

	// Check if file was found
	if fileBytes == nil {
		r.logger.Warn("File not found in local storage", zap.String("fileID", id.Hex()))
		return nil, nil
	}

	// Deserialize the file
	file, err := dom_file.NewFromDeserialized(fileBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize file", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize file", err)
	}

	return file, nil
}
