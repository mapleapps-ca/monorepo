// internal/service/sync/full.go
package sync

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	"go.uber.org/zap"
)

// FullSyncInput represents input for full synchronization
type FullSyncInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`
	MaxBatches int   `json:"max_batches,omitempty"`
}

// SyncFullService defines the interface for full synchronization operations
type SyncFullService interface {
	// Execute performs full synchronization operations on collections
	Execute(ctx context.Context, input *FullSyncInput) (*syncdto.SyncResult, error)
}

// syncFullService implements the SyncFullService interface
type syncFullService struct {
	logger                *zap.Logger
	syncCollectionService SyncCollectionService
	syncFileService       SyncFileService
}

// NewSyncFullService creates a new sync full service
func NewSyncFullService(
	logger *zap.Logger,
	syncCollectionService SyncCollectionService,
	syncFileService SyncFileService,
) SyncFullService {
	return &syncFullService{
		logger:                logger,
		syncCollectionService: syncCollectionService,
		syncFileService:       syncFileService,
	}
}

// FullSync performs both collection and file synchronization
func (s *syncFullService) Execute(ctx context.Context, input *FullSyncInput) (*syncdto.SyncResult, error) {
	s.logger.Info("Starting full synchronization")

	// Set defaults
	if input == nil {
		input = &FullSyncInput{}
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100
	}

	s.logger.Debug("Full sync input parameters",
		zap.Int("batch_size", int(input.BatchSize)),   // Cast to int
		zap.Int("max_batches", int(input.MaxBatches))) // Cast to int

	collectionInput := &SyncCollectionsInput{
		BatchSize:  input.BatchSize,
		MaxBatches: input.MaxBatches,
	}
	if _, err := s.syncCollectionService.Execute(ctx, collectionInput); err != nil {
		return nil, err
	}

	fileInput := &SyncFilesInput{
		BatchSize:  input.BatchSize,
		MaxBatches: input.MaxBatches,
	}
	if _, err := s.syncFileService.Execute(ctx, fileInput); err != nil {
		return nil, err
	}

	return nil, nil
}
