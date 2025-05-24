// internal/usecase/localfile/cleanup_orphaned_files.go
package localfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CleanupResult represents the result of cleanup operations
type CleanupResult struct {
	TotalFilesChecked    int                  `json:"total_files_checked"`
	OrphanedFilesFound   int                  `json:"orphaned_files_found"`
	OrphanedFilesDeleted int                  `json:"orphaned_files_deleted"`
	OrphanedFileIDs      []primitive.ObjectID `json:"orphaned_file_ids"`
	Errors               []string             `json:"errors,omitempty"`
}

// CleanupOrphanedFilesUseCase defines the interface for cleaning up orphaned files
type CleanupOrphanedFilesUseCase interface {
	Execute(ctx context.Context, collectionID primitive.ObjectID) (*CleanupResult, error)
}

// cleanupOrphanedFilesUseCase implements the CleanupOrphanedFilesUseCase interface
type cleanupOrphanedFilesUseCase struct {
	logger                      *zap.Logger
	getFilesByCollectionUseCase GetFilesByCollectionUseCase
	deleteFileUseCase           fileUseCase.DeleteFileUseCase
}

// NewCleanupOrphanedFilesUseCase creates a new use case for cleaning up orphaned files
func NewCleanupOrphanedFilesUseCase(
	logger *zap.Logger,
	getFilesByCollectionUseCase GetFilesByCollectionUseCase,
	deleteFileUseCase fileUseCase.DeleteFileUseCase,
) CleanupOrphanedFilesUseCase {
	return &cleanupOrphanedFilesUseCase{
		logger:                      logger,
		getFilesByCollectionUseCase: getFilesByCollectionUseCase,
		deleteFileUseCase:           deleteFileUseCase,
	}
}

// Execute removes files that no longer have valid references
func (uc *cleanupOrphanedFilesUseCase) Execute(
	ctx context.Context,
	collectionID primitive.ObjectID,
) (*CleanupResult, error) {
	uc.logger.Debug("Cleaning up orphaned files", zap.String("collectionID", collectionID.Hex()))

	result := &CleanupResult{
		OrphanedFileIDs: make([]primitive.ObjectID, 0),
		Errors:          make([]string, 0),
	}

	// Get all files in the collection
	files, err := uc.getFilesByCollectionUseCase.Execute(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get files for cleanup", err)
	}

	result.TotalFilesChecked = len(files)

	// For each file, check if it should be considered orphaned
	// This is a placeholder implementation - you would add your specific orphan detection logic
	for _, file := range files {
		// Example orphan detection criteria:
		// - File has no valid paths
		// - File size is 0
		// - File was created but never finalized
		isOrphaned := (file.EncryptedFilePath == "" && file.FilePath == "") ||
			(file.EncryptedFileSize == 0 && file.FileSize == 0)

		if isOrphaned {
			result.OrphanedFilesFound++
			result.OrphanedFileIDs = append(result.OrphanedFileIDs, file.ID)

			// Delete the orphaned file
			err := uc.deleteFileUseCase.Execute(ctx, file.ID)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to delete orphaned file %s: %v", file.ID.Hex(), err)
				result.Errors = append(result.Errors, errorMsg)
				uc.logger.Error("Failed to delete orphaned file",
					zap.String("fileID", file.ID.Hex()),
					zap.Error(err))
			} else {
				result.OrphanedFilesDeleted++
			}
		}
	}

	uc.logger.Info("Cleanup completed",
		zap.Int("totalChecked", result.TotalFilesChecked),
		zap.Int("orphanedFound", result.OrphanedFilesFound),
		zap.Int("orphanedDeleted", result.OrphanedFilesDeleted))

	return result, nil
}
