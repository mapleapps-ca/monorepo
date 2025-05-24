// monorepo/native/desktop/maplefile-cli/internal/repo/file/save.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) Save(ctx context.Context, file *dom_file.File) error {
	// Update modified timestamp
	file.ModifiedAt = time.Now()

	// Serialize the file
	fileBytes, err := file.Serialize()
	if err != nil {
		r.logger.Error("Failed to serialize file", zap.Error(err))
		return errors.NewAppError("failed to serialize file", err)
	}

	// Generate key for this file using the ID
	key := r.generateKey(file.ID.Hex())

	// Save to database
	if err := r.dbClient.Set(key, fileBytes); err != nil {
		r.logger.Error("Failed to save file to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save file to local storage", err)
	}

	r.logger.Debug("Successfully saved file to local storage",
		zap.String("fileID", file.ID.Hex()),
		zap.String("fileName", file.Name))
	return nil
}
