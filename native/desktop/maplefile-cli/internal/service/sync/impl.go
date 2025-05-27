// internal/service/sync/impl.go
package sync

import (
	"context"

	"go.uber.org/zap"
	// Import primitive
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

	s.logger.Debug("Collection sync input parameters",
		zap.Int("batchSize", int(input.BatchSize)),   // Cast to int
		zap.Int("maxBatches", int(input.MaxBatches))) // Cast to int

	// Get current sync state
	s.logger.Debug("Getting current sync state for collections")
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to get sync state for collections", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}
	s.logger.Debug("Successfully retrieved sync state for collections",
		zap.Time("lastCollectionSync", syncStateOutput.SyncState.LastCollectionSync),
		zap.String("lastCollectionID", syncStateOutput.SyncState.LastCollectionID.Hex())) // Convert ObjectID to string

	// Build cursor from sync state
	var cursor *syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastCollectionSync.IsZero() {
		cursor = &syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastCollectionSync,
			LastID:       syncStateOutput.SyncState.LastCollectionID,
		}
		s.logger.Debug("Using existing cursor for collection sync",
			zap.Time("lastModified", cursor.LastModified),
			zap.String("lastID", cursor.LastID.Hex())) // Convert ObjectID to string
	} else {
		s.logger.Debug("No previous sync state found for collections, starting from beginning")
	}

	// Get collections using progress service
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "collections",
		StartCursor:    cursor,
		BatchSize:      input.BatchSize,
		MaxBatches:     int(input.MaxBatches), // Cast to int
		TimeoutSeconds: 300,                   // 5 minutes
	}
	s.logger.Debug("Calling progress service for GetAllCollections",
		zap.Any("progressInput", progressInput))

	progressOutput, err := s.syncDTOProgressService.GetAllCollections(ctx, progressInput)
	if err != nil {
		s.logger.Error("Failed to get collections sync data from progress service", zap.Error(err))
		return nil, errors.NewAppError("failed to get collections sync data", err)
	}

	s.logger.Info("Received collection sync data summary",
		zap.Int("totalItems", progressOutput.TotalItems),
		zap.Int("batchesReceived", len(progressOutput.CollectionBatches)),
		zap.Any("finalCursor", progressOutput.FinalCursor)) // Assuming FinalCursor struct fields are loggable

	// Process all collection batches
	result := &syncdto.SyncResult{
		CollectionsProcessed: progressOutput.TotalItems,
	}

	// Analyze the sync data to determine what was added/updated/deleted
	// This is a simplified implementation - in a real scenario, you'd compare
	// with local data to determine the actual operations needed
	for i, batch := range progressOutput.CollectionBatches {
		s.logger.Debug("Processing collection batch",
			zap.Int("batchIndex", i),
			zap.Int("itemsInBatch", len(batch.Collections)))
		for _, collection := range batch.Collections {
			switch collection.State {
			case "active":
				result.CollectionsUpdated++
				// s.logger.Debug("Collection marked as active", zap.String("id", collection.ID.Hex())) // Optional: log each item
			case "deleted":
				result.CollectionsDeleted++
				// s.logger.Debug("Collection marked as deleted", zap.String("id", collection.ID.Hex())) // Optional: log each item
			case "":
				errorMsg := "empty collection state"
				s.logger.Warn(errorMsg, zap.String("id", collection.ID.Hex()))                  // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+collection.ID.Hex()) // Convert ObjectID to string for concatenation
			default:
				errorMsg := "unknown collection state: " + collection.State
				s.logger.Warn(errorMsg, zap.String("id", collection.ID.Hex()))                  // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+collection.ID.Hex()) // Convert ObjectID to string for concatenation
			}
		}
	}

	// Update sync state if we processed any data and got a final cursor
	if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
		saveInput := &syncstate.SaveInput{
			LastCollectionSync: &progressOutput.FinalCursor.LastModified,
			LastCollectionID:   &progressOutput.FinalCursor.LastID,
		}
		s.logger.Debug("Attempting to save sync state for collections",
			zap.Time("lastCollectionSync", *saveInput.LastCollectionSync),
			zap.String("lastCollectionID", saveInput.LastCollectionID.Hex())) // Convert ObjectID to string

		_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
		if err != nil {
			s.logger.Error("Failed to update sync state for collections", zap.Error(err))
			// Don't fail the entire operation for sync state update failure
			result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
		} else {
			s.logger.Info("Successfully updated sync state for collections")
		}
	} else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
		s.logger.Warn("Processed items but did not receive a final cursor for collections. Sync state not updated.")
	} else {
		s.logger.Info("No items processed for collections. Sync state not updated.")
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

	s.logger.Debug("File sync input parameters",
		zap.Int("batchSize", int(input.BatchSize)),   // Cast to int
		zap.Int("maxBatches", int(input.MaxBatches))) // Cast to int

	// Get current sync state
	s.logger.Debug("Getting current sync state for files")
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to get sync state for files", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}
	s.logger.Debug("Successfully retrieved sync state for files",
		zap.Time("lastFileSync", syncStateOutput.SyncState.LastFileSync),
		zap.String("lastFileID", syncStateOutput.SyncState.LastFileID.Hex())) // Convert ObjectID to string

	// Build cursor from sync state
	var cursor *syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastFileSync.IsZero() {
		cursor = &syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastFileSync,
			LastID:       syncStateOutput.SyncState.LastFileID,
		}
		s.logger.Debug("Using existing cursor for file sync",
			zap.Time("lastModified", cursor.LastModified),
			zap.String("lastID", cursor.LastID.Hex())) // Convert ObjectID to string
	} else {
		s.logger.Debug("No previous sync state found for files, starting from beginning")
	}

	// Get files using progress service
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "files",
		StartCursor:    cursor,
		BatchSize:      input.BatchSize,       // Cast to int
		MaxBatches:     int(input.MaxBatches), // Cast to int
		TimeoutSeconds: 300,                   // 5 minutes
	}
	s.logger.Debug("Calling progress service for GetAllFiles",
		zap.Any("progressInput", progressInput))

	progressOutput, err := s.syncDTOProgressService.GetAllFiles(ctx, progressInput)
	if err != nil {
		s.logger.Error("Failed to get files sync data from progress service", zap.Error(err))
		return nil, errors.NewAppError("failed to get files sync data", err)
	}

	s.logger.Info("Received file sync data summary",
		zap.Int("totalItems", progressOutput.TotalItems),
		zap.Int("batchesReceived", len(progressOutput.FileBatches)),
		zap.Any("finalCursor", progressOutput.FinalCursor)) // Assuming FinalCursor struct fields are loggable

	// Process all file batches
	result := &syncdto.SyncResult{
		FilesProcessed: progressOutput.TotalItems,
	}

	// Analyze the sync data to determine what was added/updated/deleted
	for i, batch := range progressOutput.FileBatches {
		s.logger.Debug("Processing file batch",
			zap.Int("batchIndex", i),
			zap.Int("itemsInBatch", len(batch.Files)))
		for _, file := range batch.Files {
			switch file.State {
			case "active":
				result.FilesUpdated++
				// s.logger.Debug("File marked as active", zap.String("id", file.ID.Hex())) // Optional: log each item
			case "deleted":
				result.FilesDeleted++
				// s.logger.Debug("File marked as deleted", zap.String("id", file.ID.Hex())) // Optional: log each item
			default:
				errorMsg := "unknown file state: " + file.State
				s.logger.Warn(errorMsg, zap.String("id", file.ID.Hex()))                  // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+file.ID.Hex()) // Convert ObjectID to string for concatenation
			}
		}
	}

	// Update sync state if we processed any data and got a final cursor
	if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
		saveInput := &syncstate.SaveInput{
			LastFileSync: &progressOutput.FinalCursor.LastModified,
			LastFileID:   &progressOutput.FinalCursor.LastID,
		}
		s.logger.Debug("Attempting to save sync state for files",
			zap.Time("lastFileSync", *saveInput.LastFileSync),
			zap.String("lastFileID", saveInput.LastFileID.Hex())) // Convert ObjectID to string

		_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
		if err != nil {
			s.logger.Error("Failed to update sync state for files", zap.Error(err))
			// Don't fail the entire operation for sync state update failure
			result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
		} else {
			s.logger.Info("Successfully updated sync state for files")
		}
	} else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
		s.logger.Warn("Processed items but did not receive a final cursor for files. Sync state not updated.")
	} else {
		s.logger.Info("No items processed for files. Sync state not updated.")
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

	s.logger.Debug("Full sync input parameters",
		zap.Int("collectionBatchSize", int(input.CollectionBatchSize)), // Cast to int
		zap.Int("fileBatchSize", int(input.FileBatchSize)),             // Cast to int
		zap.Int("maxBatches", int(input.MaxBatches)))                   // Cast to int

	result := &syncdto.SyncResult{}

	// Sync collections first
	s.logger.Info("Initiating collection sync as part of full sync")
	collectionSyncInput := &SyncCollectionsInput{
		BatchSize:  input.CollectionBatchSize,
		MaxBatches: input.MaxBatches,
	}

	collectionResult, err := s.SyncCollections(ctx, collectionSyncInput)
	if err != nil {
		s.logger.Error("Failed to sync collections during full sync", zap.Error(err))
		return nil, errors.NewAppError("failed to sync collections during full sync", err)
	}
	s.logger.Info("Collection sync completed as part of full sync",
		zap.Int("collections_processed", collectionResult.CollectionsProcessed),
		zap.Int("collections_updated", collectionResult.CollectionsUpdated),
		zap.Int("collections_deleted", collectionResult.CollectionsDeleted),
		zap.Int("collections_errors", len(collectionResult.Errors)))

	// Merge collection results
	result.CollectionsProcessed = collectionResult.CollectionsProcessed
	result.CollectionsAdded = collectionResult.CollectionsAdded // Assuming Add is counted in Processed/Updated
	result.CollectionsUpdated = collectionResult.CollectionsUpdated
	result.CollectionsDeleted = collectionResult.CollectionsDeleted
	result.Errors = append(result.Errors, collectionResult.Errors...)

	// Sync files
	s.logger.Info("Initiating file sync as part of full sync")
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
		s.logger.Info("File sync completed as part of full sync",
			zap.Int("files_processed", fileResult.FilesProcessed),
			zap.Int("files_updated", fileResult.FilesUpdated),
			zap.Int("files_deleted", fileResult.FilesDeleted),
			zap.Int("files_errors", len(fileResult.Errors)))
		// Merge file results
		result.FilesProcessed = fileResult.FilesProcessed
		result.FilesAdded = fileResult.FilesAdded // Assuming Add is counted in Processed/Updated
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
