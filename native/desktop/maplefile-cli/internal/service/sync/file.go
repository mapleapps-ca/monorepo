// internal/service/sync/file.go
package sync

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// SyncFilesInput represents input for syncing files
type SyncFilesInput struct {
	BatchSize  int64  `json:"batch_size,omitempty"`
	MaxBatches int    `json:"max_batches,omitempty"`
	Password   string `json:"password,omitempty"`
}

// SyncFileService defines the interface for synchronization operations
type SyncFileService interface {
	// Execute performs synchronization operations on files
	Execute(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error)
}

// syncFileService implements the SyncFileService interface
type syncFileService struct {
	logger *zap.Logger // Logger instance for structured logging.

	// Services for managing the sync state (cursor)
	syncStateGetService   syncstate.GetService
	syncStateSaveService  syncstate.SaveService
	syncStateResetService syncstate.ResetService

	// Service for fetching file metadata / data from the remote source (cloud)
	syncDTOProgressService syncdtoSvc.SyncProgressService
	syncDTOGetFilesService syncdtoSvc.GetFilesService

	// Use cases for interacting with the local file repository
	getFileUseCase uc.GetFileUseCase
}

// NewSyncFileService creates a new sync file service
func NewSyncFileService(
	logger *zap.Logger,

	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,

	syncDTOProgressService syncdtoSvc.SyncProgressService,
	syncDTOGetFilesService syncdtoSvc.GetFilesService,

	getFileUseCase uc.GetFileUseCase,
) SyncFileService {
	logger = logger.Named("SyncFileService")
	return &syncFileService{
		logger:                 logger,
		syncStateGetService:    syncStateGetService,
		syncStateSaveService:   syncStateSaveService,
		syncStateResetService:  syncStateResetService,
		syncDTOProgressService: syncDTOProgressService,
		syncDTOGetFilesService: syncDTOGetFilesService,
		getFileUseCase:         getFileUseCase,
	}
}

// SyncFiles synchronizes files from the cloud
func (s *syncFileService) Execute(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error) {
	s.logger.Info("✨ Starting file synchronization")

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
	if input.Password == "" {
		return nil, errors.NewAppError("Password is required", nil)
	}

	s.logger.Debug("⚙️ File sync input parameters",
		zap.Int("batchSize", int(input.BatchSize)),   // Cast to int
		zap.Int("maxBatches", int(input.MaxBatches))) // Cast to int

	// Get current sync state
	s.logger.Debug("⚙️ Getting current sync state for files")
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get sync state for files", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}
	s.logger.Debug("✅ Successfully retrieved sync state for files",
		zap.Time("lastFileSync", syncStateOutput.SyncState.LastFileSync),
		zap.String("lastFileID", syncStateOutput.SyncState.LastFileID.Hex())) // Convert ObjectID to string

	// Build cursor from sync state
	var cursor *syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastFileSync.IsZero() {
		cursor = &syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastFileSync,
			LastID:       syncStateOutput.SyncState.LastFileID,
		}
		s.logger.Debug("🔍 Using existing cursor for file sync",
			zap.Time("lastModified", cursor.LastModified),
			zap.String("lastID", cursor.LastID.Hex())) // Convert ObjectID to string
	} else {
		s.logger.Debug("🔍 No previous sync state found for files, starting from beginning")
	}

	// Get files using progress service
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "files",
		StartCursor:    cursor,
		BatchSize:      input.BatchSize,
		MaxBatches:     int(input.MaxBatches), // Cast to int
		TimeoutSeconds: 300,                   // 5 minutes
	}
	s.logger.Debug("⚙️ Calling progress service for GetAllFiles",
		zap.Any("progressInput", progressInput))

	progressOutput, err := s.syncDTOProgressService.GetAllFiles(ctx, progressInput)
	if err != nil {
		s.logger.Error("❌ Failed to get files sync data from progress service", zap.Error(err))
		return nil, errors.NewAppError("failed to get files sync data", err)
	}

	s.logger.Info("✅ Received file sync data summary",
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
		s.logger.Debug("⚙️ Processing file batch",
			zap.Int("batchIndex", i),
			zap.Int("itemsInBatch", len(batch.Files)))
		for _, cloudFile := range batch.Files {
			s.logger.Debug("🔍 Beginning to analyze file for syncing...",
				zap.String("id", cloudFile.ID.Hex()),
				zap.Uint64("version", cloudFile.Version),
				zap.Time("modified_at", cloudFile.ModifiedAt),
				zap.String("state", cloudFile.State),
				zap.Uint64("tombstone_version", cloudFile.TombstoneVersion),
				zap.Time("tombstone_expiry", cloudFile.TombstoneExpiry),
			)

			//
			// Get related records.
			//

			// Attempt to lookup the existing local file record using the ID from the cloud data.
			existingLocalFile, err := s.getFileUseCase.Execute(ctx, cloudFile.ID)
			if err != nil {
				// Log error if lookup fails but continue processing other items
				s.logger.Error("❌ Failed to get local file",
					zap.String("id", cloudFile.ID.Hex()),
					zap.Error(err))
				// Depending on error type, might need to handle specifically (e.g., not found vs actual DB error)
				continue // Skip processing this file if local lookup fails
			}

			//
			// CASE 1: If the local file is not found, create a new one (if not marked for deletion in cloud).
			//

			if existingLocalFile == nil {
				// For debugging purposes, log the details of the file being analyzed
				s.logger.Debug("👻 No local file found.",
					zap.String("id", cloudFile.ID.Hex()))

				// // Make sure the cloud file hasn't been deleted.
				// if cloudFile.TombstoneVersion > 0 {
				// 	s.logger.Debug("🚫 Skipping local file creation from the cloud because it has been marked for deletion in the cloud",
				// 		zap.String("id", cloudFile.ID.Hex()))
				// 	continue // Go to the next item in the loop and do not continue in this function.
				// }

				// localFile, err := s.createLocalFileFromCloudFileService.Execute(ctx, cloudFile.ID, input.Password)
				// if err != nil {
				// 	s.logger.Error("❌ Failed to get cloud file and create it locally",
				// 		zap.String("id", cloudFile.ID.Hex()),
				// 		zap.Error(err))
				// 	// Depending on error type, might need to handle specifically (e.g., not found vs actual DB error)
				// 	continue // Skip processing this file if local create fails
				// }

				// if localFile != nil {
				// 	fileSyncResult.FilesAdded++
				// }
				// continue // Go to the next item in the loop and do not continue in this function.
			}

			//
			// CASE 2: Delete locally if marked for deletion from cloud.
			//

			// // We must handle local deletion of the file.
			// if cloudFile.TombstoneVersion > existingLocalFile.Version || cloudFile.State == "deleted" {
			// 	if err := s.deleteFileUseCase.Execute(ctx, existingLocalFile.ID); err != nil {
			// 		s.logger.Error("❌ Failed to delete local file",
			// 			zap.String("file_id", existingLocalFile.ID.Hex()),
			// 			zap.Uint64("local_version", existingLocalFile.Version),
			// 			zap.Uint64("cloud_version", cloudFile.Version),
			// 			zap.Error(err))
			// 		return nil, err
			// 	}
			// 	s.logger.Debug("🗑️ Local file is marked as deleted",
			// 		zap.String("file_id", existingLocalFile.ID.Hex()),
			// 		zap.Uint64("local_version", existingLocalFile.Version),
			// 		zap.Uint64("cloud_version", cloudFile.Version))
			// 	fileSyncResult.FilesDeleted++
			// 	continue // Skip processing this file
			// }

			// //
			// // CASE 3: If the local file exists, check if it needs to be updated or deleted.
			// //
			// s.logger.Debug("🔄 Local file found, update if changes detected.",
			// 	zap.String("id", cloudFile.ID.Hex()))

			// // Local file is already same or newest version compared with the cloud file.
			// if existingLocalFile.Version >= cloudFile.Version {
			// 	s.logger.Debug("✅ Local file is already same or newest version compared with the cloud file",
			// 		zap.String("file_id", cloudFile.ID.Hex()),
			// 		zap.Uint64("local_version", existingLocalFile.Version),
			// 		zap.Uint64("cloud_version", cloudFile.Version),
			// 	)
			// 	continue // Skip processing this file
			// }

			// localFile, err := s.updateLocalFileFromCloudFileService.Execute(ctx, cloudFile.ID, input.Password)
			// if err != nil {
			// 	s.logger.Error("❌ Failed to get cloud file and save/delete it locally",
			// 		zap.String("id", cloudFile.ID.Hex()),
			// 		zap.Error(err))
			// 	// Depending on error type, might need to handle specifically (e.g., not found vs actual DB error)
			// 	continue // Skip processing this file if local create fails
			// }

			// // If localFile is not empty then it means it was updated.
			// if localFile != nil {
			// 	// For now, just incrementing updated count as a placeholder
			// 	fileSyncResult.FilesUpdated++
			// }

		}
	}

	// TODO: UNCOMMENT THE CODE BELOW WHEN THE SYNC CODE ABOVE IS COMPLETED

	// // Update sync state if we processed any data and got a final cursor
	// if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
	// 	saveInput := &syncstate.SaveInput{
	// 		LastFileSync: &progressOutput.FinalCursor.LastModified,
	// 		LastFileID:   &progressOutput.FinalCursor.LastID,
	// 	}
	// 	s.logger.Debug("⚙️ Attempting to save sync state for files",
	// 		zap.Time("lastFileSync", *saveInput.LastFileSync),
	// 		zap.String("lastFileID", saveInput.LastFileID.Hex())) // Convert ObjectID to string

	// 	_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
	// 	if err != nil {
	// 		s.logger.Error("❌ Failed to update sync state for files", zap.Error(err))
	// 		// Don't fail the entire operation for sync state update failure
	// 		result.Errors = append(result.Errors, "failed to update sync state: "+err.Error())
	// 	} else {
	// 		s.logger.Info("✅ Successfully updated sync state for files")
	// 	}
	// } else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
	// 	s.logger.Warn("⚠️ Processed items but did not receive a final cursor for files. Sync state not updated.")
	// } else {
	// 	s.logger.Info("ℹ️ No items processed for files. Sync state not updated.")
	// }

	s.logger.Info("✅ File synchronization completed",
		zap.Int("processed", result.FilesProcessed),
		zap.Int("updated", result.FilesUpdated),
		zap.Int("deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}
