// native/desktop/maplefile-cli/internal/repo/file/delete.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *fileRepository) Delete(ctx context.Context, id gocql.UUID) error {
	r.logger.Debug("Deleting file from local storage", zap.String("fileID", id.String()))

	// Generate key for this file
	key := r.generateKey(id.String())

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete file from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete file from local storage", err)
	}

	r.logger.Info("Successfully deleted file from local storage", zap.String("fileID", id.String()))
	return nil
}
