// monorepo/native/desktop/maplefile-cli/internal/repo/file/save.go
package file

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"go.uber.org/zap"
)

// Save updates an existing local file
func (r *fileRepository) Save(ctx context.Context, file *dom_file.File) error {
	r.logger.Debug("Saving file to local storage",
		zap.String("fileID", file.ID.Hex()))

	// Update modified timestamp
	file.ModifiedAt = time.Now()

	// Serialize the file metadata
	fileBytes, err := file.Serialize()
	if err != nil {
		r.logger.Error("Failed to serialize file metadata", zap.Error(err))
		return errors.NewAppError("failed to serialize file metadata", err)
	}

	// Generate key for this file
	key := r.generateKey(file.ID.Hex())

	// Save to database
	if err := r.dbClient.Set(key, fileBytes); err != nil {
		r.logger.Error("Failed to save file metadata to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save file metadata to local storage", err)
	}

	r.logger.Info("File metadata saved successfully to local storage",
		zap.String("fileID", file.ID.Hex()),
		zap.Bool("isModifiedLocally", file.IsModifiedLocally))
	return nil
}
