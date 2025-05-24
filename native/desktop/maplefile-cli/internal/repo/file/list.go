// monorepo/native/desktop/maplefile-cli/internal/repo/file/list.go
package file

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// List retrieves local files based on filter criteria
func (r *fileRepository) List(ctx context.Context, filter file.FileFilter) ([]*dom_file.File, error) {
	r.logger.Debug("Listing files from local storage",
		zap.Any("filter", filter))

	files := make([]*dom_file.File, 0)

	// Iterate through all files in the database
	err := r.dbClient.Iterate(func(key, value []byte) error {
		// Check if the key has our file prefix
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, fileKeyPrefix) {
			// Not a file key, skip
			return nil
		}

		// Deserialize the file
		file, err := dom_file.NewFromDeserialized(value)
		if err != nil {
			r.logger.Error("Failed to deserialize file while listing",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Apply filters

		// Filter by collectionID if specified
		if filter.CollectionID != nil && file.CollectionID != *filter.CollectionID {
			return nil // Skip, collection doesn't match
		}

		// Filter by sync status if specified
		if filter.SyncStatus != nil && file.SyncStatus != *filter.SyncStatus {
			return nil // Skip, sync status doesn't match
		}

		// Filter by name contains if specified
		if filter.NameContains != nil {
			if !strings.Contains(strings.ToLower(file.Name), strings.ToLower(*filter.NameContains)) {
				return nil // Skip, name doesn't contain the filter string
			}
		}

		// Filter by mime type if specified
		if filter.MimeType != nil && file.MimeType != *filter.MimeType {
			return nil // Skip, mime type doesn't match
		}

		// Add to results
		files = append(files, file)
		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through files in local storage", zap.Error(err))
		return nil, errors.NewAppError("failed to list files from local storage", err)
	}

	r.logger.Info("Successfully listed files from local storage",
		zap.Int("count", len(files)))
	return files, nil
}

// ListByCollection lists local files within a specific collection
func (r *fileRepository) ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_file.File, error) {
	r.logger.Debug("Listing files by collection from local storage",
		zap.String("collectionID", collectionID.Hex()))

	// Use the generic List method with a collection filter
	return r.List(ctx, file.FileFilter{
		CollectionID: &collectionID,
	})
}
