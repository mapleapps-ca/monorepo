// native/desktop/maplefile-cli/internal/repo/file/get_by_ids.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*dom_file.File, error) {
	if len(ids) == 0 {
		return []*dom_file.File{}, nil
	}

	r.logger.Debug("Getting files by IDs from local storage", zap.Int("count", len(ids)))

	files := make([]*dom_file.File, 0, len(ids))

	// Get each file individually
	for _, id := range ids {
		file, err := r.Get(ctx, id)
		if err != nil {
			r.logger.Error("Failed to get file by ID",
				zap.String("fileID", id.Hex()),
				zap.Error(err))
			return nil, err
		}

		// Only add non-nil files (found files)
		if file != nil {
			files = append(files, file)
		}
	}

	r.logger.Debug("Successfully retrieved files by IDs",
		zap.Int("requested", len(ids)),
		zap.Int("found", len(files)))
	return files, nil
}
