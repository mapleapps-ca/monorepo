// internal/service/sync/collection.go
package sync

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// SyncCollectionsInput represents input for syncing collections
type SyncCollectionsInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`
	MaxBatches int   `json:"max_batches,omitempty"`
}

// SyncCollectionService defines the interface for synchronization operations
type SyncCollectionService interface {
	// Execute performs synchronization operations on collections
	Execute(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error)
}

// syncCollectionService implements the SyncCollectionService interface
type syncCollectionService struct {
	logger                       *zap.Logger
	syncStateGetService          syncstate.GetService
	syncStateSaveService         syncstate.SaveService
	syncStateResetService        syncstate.ResetService
	syncDTOProgressService       syncdtoSvc.SyncProgressService
	syncDTOGetCollectionsService syncdtoSvc.GetCollectionsService
	createCollectionUseCase      uc.CreateCollectionUseCase
	getCollectionUseCase         uc.GetCollectionUseCase
	updateCollectionUseCase      uc.UpdateCollectionUseCase
	deleteCollectionUseCase      uc.DeleteCollectionUseCase
}

// NewSyncCollectionService creates a new sync collection service
func NewSyncCollectionService(
	logger *zap.Logger,
	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,
	syncDTOProgressService syncdtoSvc.SyncProgressService,
	syncDTOGetCollectionsService syncdtoSvc.GetCollectionsService,
	createCollectionUseCase uc.CreateCollectionUseCase,
	getCollectionUseCase uc.GetCollectionUseCase,
	updateCollectionUseCase uc.UpdateCollectionUseCase,
	deleteCollectionUseCase uc.DeleteCollectionUseCase,
) SyncCollectionService {
	return &syncCollectionService{
		logger:                       logger,
		syncStateGetService:          syncStateGetService,
		syncStateSaveService:         syncStateSaveService,
		syncStateResetService:        syncStateResetService,
		syncDTOProgressService:       syncDTOProgressService,
		syncDTOGetCollectionsService: syncDTOGetCollectionsService,
		createCollectionUseCase:      createCollectionUseCase,
		getCollectionUseCase:         getCollectionUseCase,
		updateCollectionUseCase:      updateCollectionUseCase,
		deleteCollectionUseCase:      deleteCollectionUseCase,
	}
}

// SyncCollections synchronizes collections from the cloud
func (s *syncCollectionService) Execute(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error) {
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
			// For debugging purpose only.
			s.logger.Debug("Beginning to analyze collection for syncing...",
				zap.String("id", collection.ID.Hex()),
				zap.Uint64("version", collection.Version),
				zap.Time("modified_at", collection.ModifiedAt),
				zap.String("state", collection.State),
				zap.Any("parent_id", collection.ParentID),
				zap.Uint64("tombstone_version", collection.TombstoneVersion),
				zap.Time("tombstone_expiry", collection.TombstoneExpiry),
			)

			// Attempt to lookup the existing local collection record.
			localCollection, err := s.getCollectionUseCase.Execute(ctx, collection.ID)
			if err != nil {
				s.logger.Error("Failed to get local collection",
					zap.String("id", collection.ID.Hex()),
					zap.Error(err))
				continue
			}
			_ = localCollection

			//TODO: HERE WE WILL ADD SYNC LOGIC.

			switch collection.State {
			case "active":
				// CASE 1: If the local collection is not found, create a new one.
				if localCollection == nil {
					// Create a new collection record.
					newCollection := &dom_collection.Collection{
						ID: collection.ID,
						// Name: collection.Name,
						// Description:     collection.Description,
						State:           collection.State,
						TombstoneExpiry: collection.TombstoneExpiry,
					}
					if err := s.createCollectionUseCase.Execute(ctx, newCollection); err != nil {
						s.logger.Error("Failed to create new collection",
							zap.String("id", collection.ID.Hex()),
							zap.Error(err))
						continue
					}
				}

				result.CollectionsUpdated++
				s.logger.Debug("Collection marked as active",
					zap.String("id", collection.ID.Hex())) // Optional: log each item
			case "deleted":
				result.CollectionsDeleted++
				s.logger.Debug("Collection marked as deleted",
					zap.String("id", collection.ID.Hex())) // Optional: log each item
			case "":
				errorMsg := "empty collection state"
				s.logger.Warn(errorMsg,
					zap.String("id", collection.ID.Hex())) // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+collection.ID.Hex()) // Convert ObjectID to string for concatenation
			default:
				errorMsg := "unknown collection state: " + collection.State
				s.logger.Warn(errorMsg,
					zap.String("id", collection.ID.Hex())) // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+collection.ID.Hex()) // Convert ObjectID to string for concatenation
			}
		}
	}

	// TODO: UNCOMMENT THE CODE BELOW WHEN THE SYNC CODE ABOVE IS COMPLETED

	// // Update sync state if we processed any data and got a final cursor
	// if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
	// 	saveInput := &syncstate.SaveInput{
	// 		LastCollectionSync: &progressOutput.FinalCursor.LastModified,
	// 		LastCollectionID:   &progressOutput.FinalCursor.LastID,
	// 	}
	// 	s.logger.Debug("Attempting to save sync state for collections",
	// 		zap.Time("lastCollectionSync", *saveInput.LastCollectionSync),
	// 		zap.String("lastCollectionID", saveInput.LastCollectionID.Hex())) // Convert ObjectID to string

	// 	_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
	// 	if err != nil {
	// 		s.logger.Error("Failed to update sync state for collections", zap.Error(err))
	// 		// Don't fail the entire operation for sync state update failure
	// 		result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
	// 	} else {
	// 		s.logger.Info("Successfully updated sync state for collections")
	// 	}
	// } else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
	// 	s.logger.Warn("Processed items but did not receive a final cursor for collections. Sync state not updated.")
	// } else {
	// 	s.logger.Info("No items processed for collections. Sync state not updated.")
	// }

	s.logger.Info("Collection synchronization completed",
		zap.Int("processed", result.CollectionsProcessed),
		zap.Int("updated", result.CollectionsUpdated),
		zap.Int("deleted", result.CollectionsDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}
