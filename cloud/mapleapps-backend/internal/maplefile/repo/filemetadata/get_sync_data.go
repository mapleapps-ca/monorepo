// cloud/backend/internal/maplefile/repo/filemetadata/get_sync_data.go
package filemetadata

import (
	"context"

	"github.com/gocql/gocql"
	dom_sync "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl fileMetadataRepositoryImpl) GetSyncData(ctx context.Context, userID gocql.UUID, cursor *dom_sync.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_sync.FileSyncResponse, error) {
	return nil, nil
}
