// internal/service/sync/file.go
package sync

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	dom_syncdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
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

	// File syncer services
	createLocalFileFromCloudFileService filesyncer.CreateLocalFileFromCloudFileService
	updateLocalFileFromCloudFileService filesyncer.UpdateLocalFileFromCloudFileService

	// Use cases for interacting with the local file repository
	getFileUseCase    uc_file.GetFileUseCase
	deleteFileUseCase uc_file.DeleteFileUseCase
}

// NewSyncFileService creates a new sync file service
func NewSyncFileService(
	logger *zap.Logger,
	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,
	syncDTOProgressService syncdtoSvc.SyncProgressService,
	createLocalFileFromCloudFileService filesyncer.CreateLocalFileFromCloudFileService,
	updateLocalFileFromCloudFileService filesyncer.UpdateLocalFileFromCloudFileService,
	getFileUseCase uc_file.GetFileUseCase,
	deleteFileUseCase uc_file.DeleteFileUseCase,
) SyncFileService {
	logger = logger.Named("SyncFileService")
	return &syncFileService{
		logger:                              logger,
		syncStateGetService:                 syncStateGetService,
		syncStateSaveService:                syncStateSaveService,
		syncStateResetService:               syncStateResetService,
		syncDTOProgressService:              syncDTOProgressService,
		createLocalFileFromCloudFileService: createLocalFileFromCloudFileService,
		updateLocalFileFromCloudFileService: updateLocalFileFromCloudFileService,
		getFileUseCase:                      getFileUseCase,
		deleteFileUseCase:                   deleteFileUseCase,
	}
}

// Execute synchronizes files from the cloud
func (s *syncFileService) Execute(ctx context.Context, input *SyncFilesInput) (*syncdto.SyncResult, error) {
	s.logger.Info("🔄 Starting file synchronization")

	// Set default input parameters if not provided
	if input == nil {
		input = &SyncFilesInput{}
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 50 // Default batch size
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100 // Default max batches
	}

	s.logger.Debug("⚙️ File sync input parameters",
		zap.Int("batchSize", int(input.BatchSize)),   // Cast to int for logging
		zap.Int("maxBatches", int(input.MaxBatches))) // Cast to int for logging

	// Retrieve the current sync state to determine the starting point for the sync
	s.logger.Debug("⏰ Getting current sync state for files")
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get sync state for files", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}
	s.logger.Debug("✅ Successfully retrieved sync state for files",
		zap.Time("lastFileSync", syncStateOutput.SyncState.LastFileSync),
		zap.String("lastFileID", syncStateOutput.SyncState.LastFileID.String())) // Convert ObjectID to string for logging

	// Build the sync cursor based on the retrieved sync state
	var currentSyncCursor *dom_syncdto.SyncCursorDTO
	if !(syncStateOutput.SyncState.LastFileSync.String() == "") {
		// If a previous sync state exists, use it to create the cursor
		currentSyncCursor = &dom_syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastFileSync,
			LastID:       syncStateOutput.SyncState.LastFileID,
		}
		s.logger.Debug("➡️ Using existing cursor for file sync",
			zap.Time("lastModified", currentSyncCursor.LastModified),
			zap.String("lastID", currentSyncCursor.LastID.String())) // Convert ObjectID to string for logging
	} else {
		// If no previous sync state exists, start syncing from the beginning (nil cursor)
		s.logger.Debug("✨ No previous sync state found for files, starting from beginning")
	}

	// Prepare input for the progress service to fetch files
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "files",               // Type of data being synced
		StartCursor:    currentSyncCursor,     // Cursor indicating where to start fetching
		BatchSize:      input.BatchSize,       // Requested batch size
		MaxBatches:     int(input.MaxBatches), // Maximum number of batches to retrieve
		TimeoutSeconds: 300,                   // Timeout for the entire fetching process (5 minutes)
	}
	s.logger.Debug("☁️ Calling progress service for GetAllFiles",
		zap.Any("progressInput", progressInput))

	// Fetch file data in batches from the remote sync service
	progressOutput, err := s.syncDTOProgressService.GetAllFiles(ctx, progressInput)
	if err != nil {
		// Add more detailed error logging
		s.logger.Error("❌ Failed to get files sync data from progress service",
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.Any("progressInput", progressInput))

		// Check if it's a specific backend error
		if strings.Contains(err.Error(), "multi-key map") {
			s.logger.Error("🔧 Backend MongoDB query error detected - check backend sort parameter construction")
			return nil, errors.NewAppError("backend database query error - contact system administrator", err)
		}

		return nil, errors.NewAppError("failed to get files sync data", err)
	}

	// Log summary of the fetched sync data
	s.logger.Info("📊 Received file sync data summary",
		zap.Int("totalItems", progressOutput.TotalItems),            // Total number of items across all batches
		zap.Int("batchesReceived", len(progressOutput.FileBatches)), // Number of batches received
		zap.Any("finalCursor", progressOutput.FinalCursor))          // The cursor to use for the next sync

	// Initialize the sync result structure
	fileSyncResult := &dom_syncdto.SyncResult{
		FilesProcessed: progressOutput.TotalItems,
	}

	// Process each batch of files received from the sync service
	for batchIndex, batch := range progressOutput.FileBatches {
		s.logger.Debug("📦 Processing file batch",
			zap.Int("batchIndex", batchIndex),
			zap.Int("itemsInBatch", len(batch.Files)))

		// Process each individual file within the current batch
		for _, cloudFile := range batch.Files {
			// Log detailed information about the file being analyzed
			s.logger.Debug("🔍 Beginning to analyze file for syncing...",
				zap.String("id", cloudFile.ID.String()),
				zap.Uint64("version", cloudFile.Version),
				zap.Time("modified_at", cloudFile.ModifiedAt),
				zap.String("state", cloudFile.State),
				zap.String("collection_id", cloudFile.CollectionID.String()),
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
					zap.String("id", cloudFile.ID.String()),
					zap.Error(err))
				fileSyncResult.Errors = append(fileSyncResult.Errors, "failed to get local file "+cloudFile.ID.String()+": "+err.Error())
				continue // Skip processing this file if local lookup fails
			}

			//
			// CASE 1: If the local file is not found, create a new one (if not marked for deletion in cloud).
			//

			if existingLocalFile == nil {
				// For debugging purposes, log the details of the file being analyzed
				s.logger.Debug("👻 No local file found.",
					zap.String("id", cloudFile.ID.String()))

				// Make sure the cloud file hasn't been deleted.
				if cloudFile.TombstoneVersion > 0 || cloudFile.State == "deleted" {
					s.logger.Debug("🚫 Skipping local file creation from the cloud because it has been marked for deletion in the cloud",
						zap.String("id", cloudFile.ID.String()))
					continue // Go to the next item in the loop and do not continue in this function.
				}

				localFile, err := s.createLocalFileFromCloudFileService.Execute(ctx, cloudFile.ID, input.Password)
				if err != nil {
					s.logger.Error("❌ Failed to get cloud file and create it locally",
						zap.String("id", cloudFile.ID.String()),
						zap.Error(err))
					fileSyncResult.Errors = append(fileSyncResult.Errors, "failed to create local file from cloud "+cloudFile.ID.String()+": "+err.Error())
					continue // Skip processing this file if local create fails
				}

				if localFile != nil {
					fileSyncResult.FilesAdded++
				}
				continue // Go to the next item in the loop and do not continue in this function.
			}

			//
			// CASE 2: Delete locally if marked for deletion from cloud.
			//

			// We must handle local deletion of the file.
			if cloudFile.TombstoneVersion > existingLocalFile.Version || cloudFile.State == "deleted" {
				if err := s.deleteFileUseCase.Execute(ctx, existingLocalFile.ID); err != nil {
					s.logger.Error("❌ Failed to delete local file",
						zap.String("file_id", existingLocalFile.ID.String()),
						zap.Uint64("local_version", existingLocalFile.Version),
						zap.Uint64("cloud_version", cloudFile.Version),
						zap.Error(err))
					fileSyncResult.Errors = append(fileSyncResult.Errors, "failed to delete local file "+existingLocalFile.ID.String()+": "+err.Error())
					continue
				}
				s.logger.Debug("🗑️ Local file is marked as deleted",
					zap.String("file_id", existingLocalFile.ID.String()),
					zap.Uint64("local_version", existingLocalFile.Version),
					zap.Uint64("cloud_version", cloudFile.Version))
				fileSyncResult.FilesDeleted++
				continue // Skip processing this file
			}

			//
			// CASE 3: If the local file exists, check if it needs to be updated.
			//
			s.logger.Debug("🔄 Local file found, update if changes detected.",
				zap.String("id", cloudFile.ID.String()))

			// Local file is already same or newest version compared with the cloud file.
			if existingLocalFile.Version >= cloudFile.Version {
				s.logger.Debug("✅ Local file is already same or newest version compared with the cloud file",
					zap.String("file_id", cloudFile.ID.String()),
					zap.Uint64("local_version", existingLocalFile.Version),
					zap.Uint64("cloud_version", cloudFile.Version),
				)
				continue // Skip processing this file
			}

			localFile, err := s.updateLocalFileFromCloudFileService.Execute(ctx, cloudFile.ID, input.Password)
			if err != nil {
				s.logger.Error("❌ Failed to get cloud file and save/delete it locally",
					zap.String("id", cloudFile.ID.String()),
					zap.Error(err))
				fileSyncResult.Errors = append(fileSyncResult.Errors, "failed to update local file from cloud "+cloudFile.ID.String()+": "+err.Error())
				continue // Skip processing this file if local update fails
			}

			// If localFile is not empty then it means it was updated.
			if localFile != nil {
				// For now, just incrementing updated count as a placeholder
				fileSyncResult.FilesUpdated++
			}
		}
	}

	// Update sync state if we processed any data and got a final cursor
	if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
		saveInput := &syncstate.SaveInput{
			LastFileSync: &progressOutput.FinalCursor.LastModified,
			LastFileID:   &progressOutput.FinalCursor.LastID,
		}
		s.logger.Debug("💾 Attempting to save sync state for files",
			zap.Time("lastFileSync", *saveInput.LastFileSync),
			zap.String("lastFileID", saveInput.LastFileID.String())) // Convert ObjectID to string for logging

		_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
		if err != nil {
			s.logger.Error("❌ Failed to update sync state for files", zap.Error(err))
			// Don't fail the entire operation for sync state update failure, just log and add to errors
			fileSyncResult.Errors = append(fileSyncResult.Errors, "failed to update sync state: "+err.Error())
		} else {
			s.logger.Info("✅ Successfully updated sync state for files")
		}
	} else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
		// This case indicates an issue where items were processed but no final cursor was provided.
		s.logger.Warn("⚠️ Processed items but did not receive a final cursor for files. Sync state not updated.")
		fileSyncResult.Errors = append(fileSyncResult.Errors, "processed items but no final sync cursor received")
	} else {
		// No items processed, likely nothing new to sync in this run.
		s.logger.Info("💤 No items processed for files. Sync state not updated.")
	}

	// Log final summary of the synchronization process
	s.logger.Info("🎉 File synchronization completed",
		zap.Int("processed", fileSyncResult.FilesProcessed), // Total items received from sync service
		zap.Int("added", fileSyncResult.FilesAdded),         // Items locally created
		zap.Int("updated", fileSyncResult.FilesUpdated),     // Items locally updated
		zap.Int("deleted", fileSyncResult.FilesDeleted),     // Items marked for local deletion
		zap.Int("errors", len(fileSyncResult.Errors)))       // Number of errors encountered during processing

	return fileSyncResult, nil
}
