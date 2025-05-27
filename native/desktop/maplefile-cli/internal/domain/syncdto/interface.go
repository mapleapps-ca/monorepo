// native/desktop/maplefile-cli/internal/domain/syncdto/interface.go
package sync

import (
	"context"
)

// SyncDTORepository defines the interface for sync operations with the cloud
type SyncDTORepository interface {
	// GetCollectionSyncDataFromCloud retrieves collection sync data from the cloud
	GetCollectionSyncDataFromCloud(ctx context.Context, cursor *SyncCursorDTO, limit int64) (*CollectionSyncResponseDTO, error)

	// GetFileSyncDataFromCloud retrieves file sync data from the cloud
	GetFileSyncDataFromCloud(ctx context.Context, cursor *SyncCursorDTO, limit int64) (*FileSyncResponseDTO, error)
}
