// native/desktop/maplefile-cli/internal/repo/file/get_by_collection.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) GetByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_file.File, error) {
	r.logger.Debug("Getting files by collection from local storage",
		zap.String("collectionID", collectionID.Hex()))

	// Use the generic List method with a collection filter
	return r.ListByCollection(ctx, collectionID)
}
