// internal/service/sync/impl.go
package sync

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
)

// syncService implements the SyncService interface
type syncService struct {
	logger                       *zap.Logger
	syncStateGetService          syncstate.GetService
	syncStateSaveService         syncstate.SaveService
	syncStateResetService        syncstate.ResetService
	syncDTOGetCollectionsService syncdtoSvc.GetCollectionsService
	syncDTOGetFilesService       syncdtoSvc.GetFilesService
	syncDTOGetFullSyncService    syncdtoSvc.GetFullSyncService
	syncDTOProgressService       syncdtoSvc.SyncProgressService
}

// NewSyncService creates a new sync service
func NewSyncService(
	logger *zap.Logger,
	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,
	syncDTOGetCollectionsService syncdtoSvc.GetCollectionsService,
	syncDTOGetFilesService syncdtoSvc.GetFilesService,
	syncDTOGetFullSyncService syncdtoSvc.GetFullSyncService,
	syncDTOProgressService syncdtoSvc.SyncProgressService,
) SyncService {
	return &syncService{
		logger:                       logger,
		syncStateGetService:          syncStateGetService,
		syncStateSaveService:         syncStateSaveService,
		syncStateResetService:        syncStateResetService,
		syncDTOGetCollectionsService: syncDTOGetCollectionsService,
		syncDTOGetFilesService:       syncDTOGetFilesService,
		syncDTOGetFullSyncService:    syncDTOGetFullSyncService,
		syncDTOProgressService:       syncDTOProgressService,
	}
}

// SyncCollections synchronizes collections from the cloud
func (s *syncService) SyncCollections(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error) {
	s.logger.Info("Starting collection synchronization")

	// Set defaults
	if input == nil {
		input = &SyncCollectionsInput{}
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100
	}

	// Get current sync state
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to get sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	// Build cursor from sync state
	var cursor *syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastCollectionSync.IsZero() {
		cursor = &syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastCollectionSync,
			LastID:       syncStateOutput.SyncState.LastCollectionID,
		}
	}

	// Get collections using progress service
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "collections",
		StartCursor:    cursor,
		BatchSize:      input.BatchSize,
		MaxBatches:     input.MaxBatches,
		TimeoutSeconds: 300, // 5 minutes
	}

	progressOutput, err := s.syncDTOProgressService.GetAllCollections(ctx, progressInput)
	if err != nil {
		s.logger.Error("Failed to get collections sync data", zap.Error(err))
		return nil, errors.NewAppError("failed to get collections sync data", err)
	}

	// Process all collection batches
	result := &syncdto.SyncResult{
		CollectionsProcessed: progressOutput.TotalItems,
	}

	// Analyze the sync data to determine what was added/updated/deleted
	// This is a simplified implementation - in a real scenario, you'd compare
	// with local data to determine the actual operations needed
	for _, batch := range progressOutput.CollectionBatches {
		for _, collection := range batch.Collections {
			switch collection.State {
			case "active":
				result.CollectionsUpdated++
			case "deleted":
				result.CollectionsDeleted++
			default:
				result.Errors = append(result.Errors, "unknown collection state: "+collection.State)
			}
		}
	}

	// Update sync state if we processed any data
	if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
		saveInput := &syncstate.SaveInput{
			LastCollectionSync: &progressOutput.FinalCursor.LastModified,
			LastCollectionID:   &progressOutput.FinalCursor.LastID,
		}

		_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
		if err != nil {
			s.logger.Error("Failed to update sync state", zap.Error(err))
			// Don't fail the entire operation for sync state update failure
			result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
		}
	}

	s.logger.Info("Collection synchronization completed",
		zap.Int("processed", result.CollectionsProcessed),
		zap.Int("updated", result.CollectionsUpdated),
		zap.Int("deleted", result.CollectionsDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

// SyncFiles synchronizes files from the cloud
func (s *syncService) SyncFiles(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error) {
	s.logger.Info("Starting file synchronization")

	// Set defaults
	if input == nil {
		input = &SyncFilesInput{}
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100
	}

	// Get current sync state
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to get sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	// Build cursor from sync state
	var cursor *syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastFileSync.IsZero() {
		cursor = &syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastFileSync,
			LastID:       syncStateOutput.SyncState.LastFileID,
		}
	}

	// Get files using progress service
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "files",
		StartCursor:    cursor,
		BatchSize:      input.BatchSize,
		MaxBatches:     input.MaxBatches,
		TimeoutSeconds: 300, // 5 minutes
	}

	progressOutput, err := s.syncDTOProgressService.GetAllFiles(ctx, progressInput)
	if err != nil {
		s.logger.Error("Failed to get files sync data", zap.Error(err))
		return nil, errors.NewAppError("failed to get files sync data", err)
	}

	// Process all file batches
	result := &syncdto.SyncResult{
		FilesProcessed: progressOutput.TotalItems,
	}

	// Analyze the sync data to determine what was added/updated/deleted
	for _, batch := range progressOutput.FileBatches {
		for _, file := range batch.Files {
			switch file.State {
			case "active":
				result.FilesUpdated++
			case "deleted":
				result.FilesDeleted++
			default:
				result.Errors = append(result.Errors, "unknown file state: "+file.State)
			}
		}
	}

	// Update sync state if we processed any data
	if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
		saveInput := &syncstate.SaveInput{
			LastFileSync: &progressOutput.FinalCursor.LastModified,
			LastFileID:   &progressOutput.FinalCursor.LastID,
		}

		_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
		if err != nil {
			s.logger.Error("Failed to update sync state", zap.Error(err))
			// Don't fail the entire operation for sync state update failure
			result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
		}
	}

	s.logger.Info("File synchronization completed",
		zap.Int("processed", result.FilesProcessed),
		zap.Int("updated", result.FilesUpdated),
		zap.Int("deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

// FullSync performs both collection and file synchronization
func (s *syncService) FullSync(ctx context.Context, input *FullSyncInput) (*syncdto.SyncResult, error) {
	s.logger.Info("Starting full synchronization")

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

	result := &syncdto.SyncResult{}

	// Sync collections first
	collectionSyncInput := &SyncCollectionsInput{
		BatchSize:  input.CollectionBatchSize,
		MaxBatches: input.MaxBatches,
	}

	collectionResult, err := s.SyncCollections(ctx, collectionSyncInput)
	if err != nil {
		s.logger.Error("Failed to sync collections during full sync", zap.Error(err))
		return nil, errors.NewAppError("failed to sync collections during full sync", err)
	}

	// Merge collection results
	result.CollectionsProcessed = collectionResult.CollectionsProcessed
	result.CollectionsAdded = collectionResult.CollectionsAdded
	result.CollectionsUpdated = collectionResult.CollectionsUpdated
	result.CollectionsDeleted = collectionResult.CollectionsDeleted
	result.Errors = append(result.Errors, collectionResult.Errors...)

	// Sync files
	fileSyncInput := &SyncFilesInput{
		BatchSize:  input.FileBatchSize,
		MaxBatches: input.MaxBatches,
	}

	fileResult, err := s.SyncFiles(ctx, fileSyncInput)
	if err != nil {
		s.logger.Error("Failed to sync files during full sync", zap.Error(err))
		// Don't fail the entire operation if files fail after collections succeed
		result.Errors = append(result.Errors, "failed to sync files: "+err.Error())
	} else {
		// Merge file results
		result.FilesProcessed = fileResult.FilesProcessed
		result.FilesAdded = fileResult.FilesAdded
		result.FilesUpdated = fileResult.FilesUpdated
		result.FilesDeleted = fileResult.FilesDeleted
		result.Errors = append(result.Errors, fileResult.Errors...)
	}

	s.logger.Info("Full synchronization completed",
		zap.Int("collections_processed", result.CollectionsProcessed),
		zap.Int("files_processed", result.FilesProcessed),
		zap.Int("total_errors", len(result.Errors)))

	return result, nil
}

// ResetSync resets the synchronization state
func (s *syncService) ResetSync(ctx context.Context) error {
	s.logger.Info("Resetting synchronization state")

	_, err := s.syncStateResetService.ResetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to reset sync state", zap.Error(err))
		return errors.NewAppError("failed to reset sync state", err)
	}

	s.logger.Info("Synchronization state reset successfully")
	return nil
}
