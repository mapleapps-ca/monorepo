// native/desktop/maplefile-cli/internal/repo/file/get_by_collection.go
package file

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) GetByCollection(ctx context.Context, collectionID gocql.UUID) ([]*dom_file.File, error) {
	r.logger.Debug("Getting files by collection from local storage",
		zap.String("collectionID", collectionID.String()))

	// Use the generic List method with a collection filter
	return r.ListByCollection(ctx, collectionID)
}
