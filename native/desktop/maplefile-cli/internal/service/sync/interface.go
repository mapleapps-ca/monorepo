// internal/service/sync/interface.go
package sync

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// SyncCollectionsInput represents input for syncing collections
type SyncCollectionsInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`
	MaxBatches int   `json:"max_batches,omitempty"`
}

// SyncFilesInput represents input for syncing files
type SyncFilesInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`
	MaxBatches int   `json:"max_batches,omitempty"`
}

// FullSyncInput represents input for full synchronization
type FullSyncInput struct {
	CollectionBatchSize int64 `json:"collection_batch_size,omitempty"`
	FileBatchSize       int64 `json:"file_batch_size,omitempty"`
	MaxBatches          int   `json:"max_batches,omitempty"`
}

// SyncService defines the interface for synchronization operations
type SyncService interface {
	// SyncCollections synchronizes collections from the cloud
	SyncCollections(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error)

	// SyncFiles synchronizes files from the cloud
	SyncFiles(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error)

	// FullSync performs both collection and file synchronization
	FullSync(ctx context.Context, input *FullSyncInput) (*syncdto.SyncResult, error)

	// ResetSync resets the synchronization state
	ResetSync(ctx context.Context) error
}
