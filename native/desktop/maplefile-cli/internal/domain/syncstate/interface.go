// native/desktop/maplefile-cli/internal/domain/syncstate/interface.go
package syncstate

import (
	"context"
)

// SyncStateRepository defines the interface for managing local sync state
type SyncStateRepository interface {
	// GetSyncState retrieves the current sync state
	GetSyncState(ctx context.Context) (*SyncState, error)

	// SaveSyncState saves the sync state
	SaveSyncState(ctx context.Context, state *SyncState) error

	// ResetSyncState resets the sync state (for initial sync)
	ResetSyncState(ctx context.Context) error
}
