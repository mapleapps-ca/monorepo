// internal/service/sync/full.go
package sync

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	"go.uber.org/zap"
)

// FullSyncInput represents input for full synchronization
type FullSyncInput struct {
	CollectionBatchSize int64  `json:"collection_batch_size,omitempty"`
	FileBatchSize       int64  `json:"file_batch_size,omitempty"`
	MaxBatches          int    `json:"max_batches,omitempty"`
	Password            string `json:"password,omitempty"`
}

// SyncFullService defines the interface for full synchronization operations
type SyncFullService interface {
	// Execute performs full synchronization operations on collections and files
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
	logger = logger.Named("SyncFullService")
	return &syncFullService{
		logger:                logger,
		syncCollectionService: syncCollectionService,
		syncFileService:       syncFileService,
	}
}

// Execute performs both collection and file synchronization
func (s *syncFullService) Execute(ctx context.Context, input *FullSyncInput) (*syncdto.SyncResult, error) {
	s.logger.Info("🚀 Starting full synchronization")

	// Set defaults
	if input == nil {
		input = &FullSyncInput{}
	}
	if input.CollectionBatchSize <= 0 {
		input.CollectionBatchSize = 50
	}
	if input.FileBatchSize <= 0 {
		input.FileBatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100
	}
	if input.Password == "" {
		s.logger.Error("❌ Password is required for full sync")
		return nil, errors.NewAppError("password is required for E2EE operations", nil)
	}

	s.logger.Debug("⚙️ Full sync input parameters",
		zap.Int("collection_batch_size", int(input.CollectionBatchSize)),
		zap.Int("file_batch_size", int(input.FileBatchSize)),
		zap.Int("max_batches", int(input.MaxBatches)))

	// Initialize combined result
	combinedResult := &syncdto.SyncResult{}

	// Step 1: Sync collections
	s.logger.Info("📁 Starting collection synchronization...")
	collectionInput := &SyncCollectionsInput{
		BatchSize:  input.CollectionBatchSize,
		MaxBatches: input.MaxBatches,
		Password:   input.Password,
	}

	collectionResult, err := s.syncCollectionService.Execute(ctx, collectionInput)
	if err != nil {
		s.logger.Error("❌ Collection sync failed", zap.Error(err))
		return nil, err
	}

	// Merge collection results
	combinedResult.CollectionsProcessed = collectionResult.CollectionsProcessed
	combinedResult.CollectionsAdded = collectionResult.CollectionsAdded
	combinedResult.CollectionsUpdated = collectionResult.CollectionsUpdated
	combinedResult.CollectionsDeleted = collectionResult.CollectionsDeleted
	combinedResult.Errors = append(combinedResult.Errors, collectionResult.Errors...)

	s.logger.Info("✅ Collection synchronization completed",
		zap.Int("processed", collectionResult.CollectionsProcessed),
		zap.Int("added", collectionResult.CollectionsAdded),
		zap.Int("updated", collectionResult.CollectionsUpdated),
		zap.Int("deleted", collectionResult.CollectionsDeleted))

	// Step 2: Sync files
	s.logger.Info("📄 Starting file synchronization...")
	fileInput := &SyncFilesInput{
		BatchSize:  input.FileBatchSize,
		MaxBatches: input.MaxBatches,
		Password:   input.Password,
	}

	fileResult, err := s.syncFileService.Execute(ctx, fileInput)
	if err != nil {
		s.logger.Error("❌ File sync failed", zap.Error(err))
		return nil, err
	}

	// Merge file results
	combinedResult.FilesProcessed = fileResult.FilesProcessed
	combinedResult.FilesAdded = fileResult.FilesAdded
	combinedResult.FilesUpdated = fileResult.FilesUpdated
	combinedResult.FilesDeleted = fileResult.FilesDeleted
	combinedResult.Errors = append(combinedResult.Errors, fileResult.Errors...)

	s.logger.Info("✅ File synchronization completed",
		zap.Int("processed", fileResult.FilesProcessed),
		zap.Int("added", fileResult.FilesAdded),
		zap.Int("updated", fileResult.FilesUpdated),
		zap.Int("deleted", fileResult.FilesDeleted))

	// Log final summary
	totalProcessed := combinedResult.CollectionsProcessed + combinedResult.FilesProcessed
	totalModified := combinedResult.CollectionsAdded + combinedResult.CollectionsUpdated + combinedResult.CollectionsDeleted +
		combinedResult.FilesAdded + combinedResult.FilesUpdated + combinedResult.FilesDeleted

	s.logger.Info("🎉 Full synchronization completed",
		zap.Int("total_processed", totalProcessed),
		zap.Int("total_modified", totalModified),
		zap.Int("total_errors", len(combinedResult.Errors)))

	return combinedResult, nil
}
