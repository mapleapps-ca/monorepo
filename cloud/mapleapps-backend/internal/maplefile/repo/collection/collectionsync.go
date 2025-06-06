// cloud/backend/internal/maplefile/repo/collection/collectionsync.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	dom_sync "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) GetCollectionSyncData(ctx context.Context, userID gocql.UUID, cursor *dom_sync.CollectionSyncCursor, limit int64) (*dom_sync.CollectionSyncResponse, error) {
	return nil, nil //TODO
}
