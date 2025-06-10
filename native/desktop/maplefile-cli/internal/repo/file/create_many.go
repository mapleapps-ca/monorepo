// monorepo/native/desktop/maplefile-cli/internal/repo/file/create_many.go
package file

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) CreateMany(ctx context.Context, files []*dom_file.File) error {
	if len(files) == 0 {
		return nil
	}

	r.logger.Debug("Creating multiple files in local storage", zap.Int("count", len(files)))

	// Set IDs and timestamps for all files
	now := time.Now()
	for _, file := range files {
		if file.ID.String() == "" {
			file.ID = gocql.TimeUUID()
		}
		if file.CreatedAt.String() == "" {
			file.CreatedAt = now
		}
		file.ModifiedAt = now
	}

	// Save each file (could be optimized with batch operations if LevelDB supports it)
	for _, file := range files {
		if err := r.Save(ctx, file); err != nil {
			r.logger.Error("Failed to save file during batch create",
				zap.String("fileID", file.ID.String()),
				zap.Error(err))
			return errors.NewAppError("failed to save file during batch create", err)
		}
	}

	r.logger.Info("Successfully created multiple files", zap.Int("count", len(files)))
	return nil
}
