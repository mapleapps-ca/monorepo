// native/desktop/maplefile-cli/internal/repo/file/delete_many.go
package file

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (r *fileRepository) DeleteMany(ctx context.Context, ids []gocql.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	r.logger.Debug("Deleting multiple files from local storage", zap.Int("count", len(ids)))

	// Delete each file individually
	for _, id := range ids {
		if err := r.Delete(ctx, id); err != nil {
			r.logger.Error("Failed to delete file during batch delete",
				zap.String("fileID", id.String()),
				zap.Error(err))
			return err
		}
	}

	r.logger.Info("Successfully deleted multiple files", zap.Int("count", len(ids)))
	return nil
}
