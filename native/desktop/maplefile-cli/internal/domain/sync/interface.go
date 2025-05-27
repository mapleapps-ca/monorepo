// native/desktop/maplefile-cli/internal/domain/sync/interface.go
package sync

import (
	"context"
)

// SyncRepository defines the interface for sync operations with the cloud
type SyncRepository interface {
	// GetCollectionSyncData retrieves collection sync data from the cloud
	GetCollectionSyncData(ctx context.Context, cursor *SyncCursor, limit int64) (*CollectionSyncResponse, error)

	// GetFileSyncData retrieves file sync data from the cloud
	GetFileSyncData(ctx context.Context, cursor *SyncCursor, limit int64) (*FileSyncResponse, error)
}

// SyncStateRepository defines the interface for managing local sync state
type SyncStateRepository interface {
	// GetSyncState retrieves the current sync state
	GetSyncState(ctx context.Context) (*SyncState, error)

	// SaveSyncState saves the sync state
	SaveSyncState(ctx context.Context, state *SyncState) error

	// ResetSyncState resets the sync state (for initial sync)
	ResetSyncState(ctx context.Context) error
}
