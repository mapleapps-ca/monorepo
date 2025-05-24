// monorepo/native/desktop/maplefile-cli/internal/repo/file/check_by_id.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *fileRepository) CheckIfExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	r.logger.Debug("Checking if file exists by ID", zap.String("fileID", id.Hex()))

	// Generate key for this file
	key := r.generateKey(id.Hex())

	// Get from database
	fileBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to check file existence in local storage",
			zap.String("key", key),
			zap.Error(err))
		return false, errors.NewAppError("failed to check file existence in local storage", err)
	}

	// File exists if we got non-nil data
	exists := fileBytes != nil
	r.logger.Debug("File existence check completed",
		zap.String("fileID", id.Hex()),
		zap.Bool("exists", exists))

	return exists, nil
}
