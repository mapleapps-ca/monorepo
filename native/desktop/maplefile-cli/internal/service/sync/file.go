// internal/service/sync/file.go
package sync

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	"go.uber.org/zap"
)

// SyncFilesInput represents input for syncing files
type SyncFilesInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`
	MaxBatches int   `json:"max_batches,omitempty"`
}

// SyncFileService defines the interface for synchronization operations
type SyncFileService interface {
	// Execute performs synchronization operations on files
	Execute(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error)
}

// syncFileService implements the SyncFileService interface
type syncFileService struct {
	logger                 *zap.Logger
	syncStateGetService    syncstate.GetService
	syncStateSaveService   syncstate.SaveService
	syncStateResetService  syncstate.ResetService
	syncDTOProgressService syncdtoSvc.SyncProgressService
	syncDTOGetFilesService syncdtoSvc.GetFilesService
}

// NewSyncFileService creates a new sync file service
func NewSyncFileService(
	logger *zap.Logger,
	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,
	syncDTOProgressService syncdtoSvc.SyncProgressService,
	syncDTOGetFilesService syncdtoSvc.GetFilesService,
) SyncFileService {
	logger = logger.Named("SyncFileService")
	return &syncFileService{
		logger:                 logger,
		syncStateGetService:    syncStateGetService,
		syncStateSaveService:   syncStateSaveService,
		syncStateResetService:  syncStateResetService,
		syncDTOProgressService: syncDTOProgressService,
		syncDTOGetFilesService: syncDTOGetFilesService,
	}
}

// SyncFiles synchronizes files from the cloud
func (s *syncFileService) Execute(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error) {
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
		BatchSize:      input.BatchSize,
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
	// This is a simplified implementation - in a real scenario, you'd compare
	// with local data to determine the actual operations needed
	for i, batch := range progressOutput.FileBatches {
		s.logger.Debug("Processing file batch",
			zap.Int("batchIndex", i),
			zap.Int("itemsInBatch", len(batch.Files)))
		for _, file := range batch.Files {
			s.logger.Debug("Beginning to analyze file for syncing...",
				zap.String("id", file.ID.Hex()),
				zap.Uint64("version", file.Version),
				zap.Time("modified_at", file.ModifiedAt),
				zap.String("state", file.State),
				zap.Uint64("tombstone_version", file.TombstoneVersion),
				zap.Time("tombstone_expiry", file.TombstoneExpiry),
			)

			//TODO: HERE WE WILL ADD SYNC LOGIC.

			switch file.State {
			case "active":
				result.FilesUpdated++
				s.logger.Debug("File marked as active",
					zap.String("id", file.ID.Hex())) // Optional: log each item
			case "deleted":
				result.FilesDeleted++
				s.logger.Debug("File marked as deleted",
					zap.String("id", file.ID.Hex())) // Optional: log each item
			case "":
				errorMsg := "empty file state"
				s.logger.Warn(errorMsg,
					zap.String("id", file.ID.Hex())) // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+file.ID.Hex()) // Convert ObjectID to string for concatenation
			default:
				errorMsg := "unknown file state: " + file.State
				s.logger.Warn(errorMsg,
					zap.String("id", file.ID.Hex())) // Convert ObjectID to string
				result.Errors = append(result.Errors, errorMsg+" for ID: "+file.ID.Hex()) // Convert ObjectID to string for concatenation
			}
		}
	}

	// TODO: UNCOMMENT THE CODE BELOW WHEN THE SYNC CODE ABOVE IS COMPLETED

	// // Update sync state if we processed any data and got a final cursor
	// if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
	// 	saveInput := &syncstate.SaveInput{
	// 		LastFileSync: &progressOutput.FinalCursor.LastModified,
	// 		LastFileID:   &progressOutput.FinalCursor.LastID,
	// 	}
	// 	s.logger.Debug("Attempting to save sync state for files",
	// 		zap.Time("lastFileSync", *saveInput.LastFileSync),
	// 		zap.String("lastFileID", saveInput.LastFileID.Hex())) // Convert ObjectID to string

	// 	_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
	// 	if err != nil {
	// 		s.logger.Error("Failed to update sync state for files", zap.Error(err))
	// 		// Don't fail the entire operation for sync state update failure
	// 		result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
	// 	} else {
	// 		s.logger.Info("Successfully updated sync state for files")
	// 	}
	// } else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
	// 	s.logger.Warn("Processed items but did not receive a final cursor for files. Sync state not updated.")
	// } else {
	// 	s.logger.Info("No items processed for files. Sync state not updated.")
	// }

	s.logger.Info("File synchronization completed",
		zap.Int("processed", result.FilesProcessed),
		zap.Int("updated", result.FilesUpdated),
		zap.Int("deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}
