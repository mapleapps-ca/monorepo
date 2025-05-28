// internal/usecase/syncdto/process_sync_response.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// ProcessSyncResponseInput represents the input for processing sync responses
type ProcessSyncResponseInput struct {
	CollectionResponse *syncdto.CollectionSyncResponseDTO
	FileResponse       *syncdto.FileSyncResponseDTO
}

// ProcessSyncResponseOutput represents the output of processing sync responses
type ProcessSyncResponseOutput struct {
	Result               *syncdto.SyncResult
	NextCollectionCursor *syncdto.SyncCursorDTO
	NextFileCursor       *syncdto.SyncCursorDTO
	HasMoreCollections   bool
	HasMoreFiles         bool
}

// ProcessSyncResponseUseCase defines the interface for processing sync responses
type ProcessSyncResponseUseCase interface {
	Execute(ctx context.Context, input *ProcessSyncResponseInput) (*ProcessSyncResponseOutput, error)
	ProcessCollectionResponse(ctx context.Context, response *syncdto.CollectionSyncResponseDTO) (*syncdto.SyncResult, error)
	ProcessFileResponse(ctx context.Context, response *syncdto.FileSyncResponseDTO) (*syncdto.SyncResult, error)
}

// processSyncResponseUseCase implements the ProcessSyncResponseUseCase interface
type processSyncResponseUseCase struct {
	logger                 *zap.Logger
	buildSyncCursorUseCase BuildSyncCursorUseCase
}

// NewProcessSyncResponseUseCase creates a new use case for processing sync responses
func NewProcessSyncResponseUseCase(
	logger *zap.Logger,
	buildSyncCursorUseCase BuildSyncCursorUseCase,
) ProcessSyncResponseUseCase {
	logger = logger.Named("ProcessSyncResponseUseCase")
	return &processSyncResponseUseCase{
		logger:                 logger,
		buildSyncCursorUseCase: buildSyncCursorUseCase,
	}
}

// Execute processes both collection and file sync responses
func (uc *processSyncResponseUseCase) Execute(ctx context.Context, input *ProcessSyncResponseInput) (*ProcessSyncResponseOutput, error) {
	// Validate input
	if input == nil {
		return nil, errors.NewAppError("process sync response input is required", nil)
	}

	uc.logger.Debug("Processing sync responses")

	output := &ProcessSyncResponseOutput{
		Result: &syncdto.SyncResult{},
	}

	// Process collection response if provided
	if input.CollectionResponse != nil {
		collectionResult, err := uc.ProcessCollectionResponse(ctx, input.CollectionResponse)
		if err != nil {
			return nil, errors.NewAppError("failed to process collection response", err)
		}

		// Merge collection results
		output.Result.CollectionsProcessed = collectionResult.CollectionsProcessed
		output.Result.CollectionsAdded = collectionResult.CollectionsAdded
		output.Result.CollectionsUpdated = collectionResult.CollectionsUpdated
		output.Result.CollectionsDeleted = collectionResult.CollectionsDeleted
		output.Result.Errors = append(output.Result.Errors, collectionResult.Errors...)

		// Set collection cursor and hasMore flag
		output.NextCollectionCursor = input.CollectionResponse.NextCursor
		output.HasMoreCollections = input.CollectionResponse.HasMore
	}

	// Process file response if provided
	if input.FileResponse != nil {
		fileResult, err := uc.ProcessFileResponse(ctx, input.FileResponse)
		if err != nil {
			return nil, errors.NewAppError("failed to process file response", err)
		}

		// Merge file results
		output.Result.FilesProcessed = fileResult.FilesProcessed
		output.Result.FilesAdded = fileResult.FilesAdded
		output.Result.FilesUpdated = fileResult.FilesUpdated
		output.Result.FilesDeleted = fileResult.FilesDeleted
		output.Result.Errors = append(output.Result.Errors, fileResult.Errors...)

		// Set file cursor and hasMore flag
		output.NextFileCursor = input.FileResponse.NextCursor
		output.HasMoreFiles = input.FileResponse.HasMore
	}

	uc.logger.Debug("Successfully processed sync responses",
		zap.Int("collectionsProcessed", output.Result.CollectionsProcessed),
		zap.Int("filesProcessed", output.Result.FilesProcessed))

	return output, nil
}

// ProcessCollectionResponse processes a collection sync response
func (uc *processSyncResponseUseCase) ProcessCollectionResponse(ctx context.Context, response *syncdto.CollectionSyncResponseDTO) (*syncdto.SyncResult, error) {
	// Validate input
	if response == nil {
		return nil, errors.NewAppError("collection sync response is required", nil)
	}

	uc.logger.Debug("Processing collection sync response",
		zap.Int("collectionsCount", len(response.Collections)))

	result := &syncdto.SyncResult{}

	// Process each collection item
	for _, collection := range response.Collections {
		result.CollectionsProcessed++

		// Determine action based on state
		switch collection.State {
		case "active":
			// This could be new or updated - would need additional logic to determine
			// For now, assume all active items are updates
			result.CollectionsUpdated++
		case "deleted":
			result.CollectionsDeleted++
		default:
			// Unknown state, count as error
			result.Errors = append(result.Errors, "unknown collection state: "+collection.State)
		}
	}

	uc.logger.Debug("Collection response processed",
		zap.Int("processed", result.CollectionsProcessed),
		zap.Int("updated", result.CollectionsUpdated),
		zap.Int("deleted", result.CollectionsDeleted))

	return result, nil
}

// ProcessFileResponse processes a file sync response
func (uc *processSyncResponseUseCase) ProcessFileResponse(ctx context.Context, response *syncdto.FileSyncResponseDTO) (*syncdto.SyncResult, error) {
	// Validate input
	if response == nil {
		return nil, errors.NewAppError("file sync response is required", nil)
	}

	uc.logger.Debug("Processing file sync response",
		zap.Int("filesCount", len(response.Files)))

	result := &syncdto.SyncResult{}

	// Process each file item
	for _, file := range response.Files {
		result.FilesProcessed++

		// Determine action based on state
		switch file.State {
		case "active":
			// This could be new or updated - would need additional logic to determine
			// For now, assume all active items are updates
			result.FilesUpdated++
		case "deleted":
			result.FilesDeleted++
		default:
			// Unknown state, count as error
			result.Errors = append(result.Errors, "unknown file state: "+file.State)
		}
	}

	uc.logger.Debug("File response processed",
		zap.Int("processed", result.FilesProcessed),
		zap.Int("updated", result.FilesUpdated),
		zap.Int("deleted", result.FilesDeleted))

	return result, nil
}
