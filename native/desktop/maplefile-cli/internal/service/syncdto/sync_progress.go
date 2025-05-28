// internal/service/syncdto/sync_progress.go
package syncdto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyncProgressInput represents the input for managing sync progress
type SyncProgressInput struct {
	SyncType       string                 `json:"sync_type"` // "collections" or "files"
	StartCursor    *syncdto.SyncCursorDTO `json:"start_cursor,omitempty"`
	BatchSize      int64                  `json:"batch_size,omitempty"`
	MaxBatches     int                    `json:"max_batches,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
}

// SyncProgressOutput represents the result of sync progress operations
type SyncProgressOutput struct {
	SyncType          string                               `json:"sync_type"`
	TotalBatches      int                                  `json:"total_batches"`
	ProcessedBatches  int                                  `json:"processed_batches"`
	TotalItems        int                                  `json:"total_items"`
	CollectionBatches []*syncdto.CollectionSyncResponseDTO `json:"collection_batches,omitempty"`
	FileBatches       []*syncdto.FileSyncResponseDTO       `json:"file_batches,omitempty"`
	FinalCursor       *syncdto.SyncCursorDTO               `json:"final_cursor,omitempty"`
	HasMoreData       bool                                 `json:"has_more_data"`
	ElapsedTime       time.Duration                        `json:"elapsed_time"`
	Message           string                               `json:"message"`
}

// SyncProgressService defines the interface for managing paginated sync operations
type SyncProgressService interface {
	GetAllCollections(ctx context.Context, input *SyncProgressInput) (*SyncProgressOutput, error)
	GetAllFiles(ctx context.Context, input *SyncProgressInput) (*SyncProgressOutput, error)
	GetIncrementalSync(ctx context.Context, lastModified time.Time, lastID primitive.ObjectID, syncType string) (*SyncProgressOutput, error)
}

// syncProgressService implements the SyncProgressService interface
type syncProgressService struct {
	logger      *zap.Logger
	syncDTORepo syncdto.SyncDTORepository
}

// NewSyncProgressService creates a new service for managing sync progress
func NewSyncProgressService(
	logger *zap.Logger,
	syncDTORepo syncdto.SyncDTORepository,
) SyncProgressService {
	logger = logger.Named("SyncProgressService")
	return &syncProgressService{
		logger:      logger,
		syncDTORepo: syncDTORepo,
	}
}

// GetAllCollections retrieves all collections in batches with progress tracking
func (s *syncProgressService) GetAllCollections(ctx context.Context, input *SyncProgressInput) (*SyncProgressOutput, error) {
	if input == nil {
		input = &SyncProgressInput{}
	}

	// Set defaults
	if input.BatchSize <= 0 {
		input.BatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100 // Prevent infinite loops
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 300 // 5 minutes default
	}

	startTime := time.Now()
	timeout := time.Duration(input.TimeoutSeconds) * time.Second

	s.logger.Info("Starting paginated collection sync",
		zap.Int64("batch_size", input.BatchSize),
		zap.Int("max_batches", input.MaxBatches),
		zap.Duration("timeout", timeout))

	output := &SyncProgressOutput{
		SyncType:          "collections",
		CollectionBatches: make([]*syncdto.CollectionSyncResponseDTO, 0),
	}

	currentCursor := input.StartCursor
	batchCount := 0

	for batchCount < input.MaxBatches {
		// Check timeout
		if time.Since(startTime) > timeout {
			s.logger.Warn("Sync operation timed out", zap.Duration("elapsed", time.Since(startTime)))
			break
		}

		// Get next batch
		response, err := s.syncDTORepo.GetCollectionSyncDataFromCloud(ctx, currentCursor, input.BatchSize)
		if err != nil {
			s.logger.Error("failed to get collection batch",
				zap.Int("batch", batchCount+1),
				zap.Error(err))
			return nil, errors.NewAppError("failed to get collection batch", err)
		}

		// Add batch to results
		output.CollectionBatches = append(output.CollectionBatches, response)
		output.TotalItems += len(response.Collections)
		batchCount++

		s.logger.Debug("Processed collection batch",
			zap.Int("batch_number", batchCount),
			zap.Int("items_in_batch", len(response.Collections)),
			zap.Int("total_items", output.TotalItems))

		// Check if we have more data
		if !response.HasMore || response.NextCursor == nil {
			s.logger.Info("No more collection data available")
			break
		}

		// Update cursor for next batch
		currentCursor = response.NextCursor
		output.FinalCursor = response.NextCursor
		output.HasMoreData = response.HasMore
	}

	output.TotalBatches = batchCount
	output.ProcessedBatches = batchCount
	output.ElapsedTime = time.Since(startTime)

	if output.TotalItems == 0 {
		output.Message = "No collection changes found"
	} else {
		output.Message = "Collection sync completed successfully"
	}

	s.logger.Info("Completed paginated collection sync",
		zap.Int("total_batches", output.TotalBatches),
		zap.Int("total_items", output.TotalItems),
		zap.Duration("elapsed_time", output.ElapsedTime))

	return output, nil
}

// GetAllFiles retrieves all files in batches with progress tracking
func (s *syncProgressService) GetAllFiles(ctx context.Context, input *SyncProgressInput) (*SyncProgressOutput, error) {
	if input == nil {
		input = &SyncProgressInput{}
	}

	// Set defaults
	if input.BatchSize <= 0 {
		input.BatchSize = 50
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 300
	}

	startTime := time.Now()
	timeout := time.Duration(input.TimeoutSeconds) * time.Second

	s.logger.Info("Starting paginated file sync",
		zap.Int64("batch_size", input.BatchSize),
		zap.Int("max_batches", input.MaxBatches),
		zap.Duration("timeout", timeout))

	output := &SyncProgressOutput{
		SyncType:    "files",
		FileBatches: make([]*syncdto.FileSyncResponseDTO, 0),
	}

	currentCursor := input.StartCursor
	batchCount := 0

	for batchCount < input.MaxBatches {
		// Check timeout
		if time.Since(startTime) > timeout {
			s.logger.Warn("Sync operation timed out", zap.Duration("elapsed", time.Since(startTime)))
			break
		}

		// Get next batch
		response, err := s.syncDTORepo.GetFileSyncDataFromCloud(ctx, currentCursor, input.BatchSize)
		if err != nil {
			s.logger.Error("failed to get file batch",
				zap.Int("batch", batchCount+1),
				zap.Error(err))
			return nil, errors.NewAppError("failed to get file batch", err)
		}

		// Add batch to results
		output.FileBatches = append(output.FileBatches, response)
		output.TotalItems += len(response.Files)
		batchCount++

		s.logger.Debug("Processed file batch",
			zap.Int("batch_number", batchCount),
			zap.Int("items_in_batch", len(response.Files)),
			zap.Int("total_items", output.TotalItems))

		// Check if we have more data
		if !response.HasMore || response.NextCursor == nil {
			s.logger.Info("No more file data available")
			break
		}

		// Update cursor for next batch
		currentCursor = response.NextCursor
		output.FinalCursor = response.NextCursor
		output.HasMoreData = response.HasMore
	}

	output.TotalBatches = batchCount
	output.ProcessedBatches = batchCount
	output.ElapsedTime = time.Since(startTime)

	if output.TotalItems == 0 {
		output.Message = "No file changes found"
	} else {
		output.Message = "File sync completed successfully"
	}

	s.logger.Info("Completed paginated file sync",
		zap.Int("total_batches", output.TotalBatches),
		zap.Int("total_items", output.TotalItems),
		zap.Duration("elapsed_time", output.ElapsedTime))

	return output, nil
}

// GetIncrementalSync performs incremental sync based on last sync timestamp
func (s *syncProgressService) GetIncrementalSync(ctx context.Context, lastModified time.Time, lastID primitive.ObjectID, syncType string) (*SyncProgressOutput, error) {
	s.logger.Info("Starting incremental sync",
		zap.Time("last_modified", lastModified),
		zap.String("last_id", lastID.Hex()),
		zap.String("sync_type", syncType))

	cursor := &syncdto.SyncCursorDTO{
		LastModified: lastModified,
		LastID:       lastID,
	}

	input := &SyncProgressInput{
		SyncType:    syncType,
		StartCursor: cursor,
		BatchSize:   100, // Larger batch for incremental
		MaxBatches:  50,  // Reasonable limit for incremental
	}

	switch syncType {
	case "collections":
		return s.GetAllCollections(ctx, input)
	case "files":
		return s.GetAllFiles(ctx, input)
	default:
		s.logger.Error("invalid sync type", zap.String("sync_type", syncType))
		return nil, errors.NewAppError("sync_type must be 'collections' or 'files'", nil)
	}
}
