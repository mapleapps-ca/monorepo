// internal/service/syncdto/sync_progress.go
package syncdto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
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
	GetIncrementalSync(ctx context.Context, lastModified time.Time, lastID gocql.UUID, syncType string) (*SyncProgressOutput, error)
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

	s.logger.Info("✨ Starting paginated collection sync",
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
			s.logger.Warn("⏱️ Sync operation timed out", zap.Duration("elapsed", time.Since(startTime)))
			break
		}

		// Get next batch
		response, err := s.syncDTORepo.GetCollectionSyncDataFromCloud(ctx, currentCursor, input.BatchSize)
		if err != nil {
			s.logger.Error("❌ failed to get collection batch",
				zap.Int("batch", batchCount+1),
				zap.Error(err))
			return nil, errors.NewAppError("failed to get collection batch", err)
		}

		// Add batch to results
		output.CollectionBatches = append(output.CollectionBatches, response)
		output.TotalItems += len(response.Collections)
		batchCount++

		s.logger.Debug("✅ Processed collection batch",
			zap.Int("batch_number", batchCount),
			zap.Int("items_in_batch", len(response.Collections)),
			zap.Int("total_items", output.TotalItems))

		// Always capture the cursor from response if available (same logic as files)
		if response.NextCursor != nil {
			output.FinalCursor = response.NextCursor
		} else if len(response.Collections) > 0 {
			lastItem := response.Collections[len(response.Collections)-1]
			output.FinalCursor = &syncdto.SyncCursorDTO{
				LastModified: lastItem.ModifiedAt,
				LastID:       lastItem.ID,
			}
		}

		// Update hasMoreData flag
		output.HasMoreData = response.HasMore

		// Check if we should continue
		if !response.HasMore {
			s.logger.Info("🏁 No more collection data available")
			break
		}

		// Update cursor for next batch
		currentCursor = response.NextCursor
		if currentCursor == nil {
			s.logger.Warn("⚠️ No next cursor provided but hasMore=true, stopping")
			break
		}

	}

	output.TotalBatches = batchCount
	output.ProcessedBatches = batchCount
	output.ElapsedTime = time.Since(startTime)

	if output.TotalItems == 0 {
		output.Message = "No collection changes found"
	} else {
		output.Message = "Collection sync completed successfully"
	}

	s.logger.Info("✅ Completed paginated collection sync",
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

	s.logger.Info("✨ Starting paginated file sync",
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
			s.logger.Warn("⏱️ Sync operation timed out", zap.Duration("elapsed", time.Since(startTime)))
			break
		}

		// Get next batch
		response, err := s.syncDTORepo.GetFileSyncDataFromCloud(ctx, currentCursor, input.BatchSize)
		if err != nil {
			s.logger.Error("❌ failed to get file batch",
				zap.Int("batch", batchCount+1),
				zap.Error(err))
			return nil, errors.NewAppError("failed to get file batch", err)
		}

		// Add batch to results
		output.FileBatches = append(output.FileBatches, response)
		output.TotalItems += len(response.Files)
		batchCount++

		s.logger.Debug("✅ Processed file batch",
			zap.Int("batch_number", batchCount),
			zap.Int("items_in_batch", len(response.Files)),
			zap.Int("total_items", output.TotalItems))

		// ✅ FIX: Always capture the cursor from response if available
		if response.NextCursor != nil {
			output.FinalCursor = response.NextCursor
			s.logger.Debug("📍 Captured cursor from response",
				zap.String("lastID", response.NextCursor.LastID.Hex()),
				zap.Time("lastModified", response.NextCursor.LastModified),
				zap.Bool("hasMore", response.HasMore))
		} else if len(response.Files) > 0 {
			// ✅ FALLBACK: Build cursor from last processed item if none provided
			lastItem := response.Files[len(response.Files)-1]
			output.FinalCursor = &syncdto.SyncCursorDTO{
				LastModified: lastItem.ModifiedAt,
				LastID:       lastItem.ID,
			}
			s.logger.Debug("📍 Built cursor from last item",
				zap.String("lastID", lastItem.ID.Hex()),
				zap.Time("lastModified", lastItem.ModifiedAt))
		}

		// Update hasMoreData flag
		output.HasMoreData = response.HasMore

		// Check if we should continue
		if !response.HasMore {
			s.logger.Info("🏁 No more file data available (hasMore=false)")
			break
		}

		// Update cursor for next batch
		currentCursor = response.NextCursor
		if currentCursor == nil {
			s.logger.Warn("⚠️ No next cursor provided but hasMore=true, stopping")
			break
		}
	}

	output.TotalBatches = batchCount
	output.ProcessedBatches = batchCount
	output.ElapsedTime = time.Since(startTime)

	if output.TotalItems == 0 {
		output.Message = "No file changes found"
	} else {
		output.Message = "File sync completed successfully"
	}

	s.logger.Info("✅ Completed paginated file sync",
		zap.Int("total_batches", output.TotalBatches),
		zap.Int("total_items", output.TotalItems),
		zap.Bool("hasFinalCursor", output.FinalCursor != nil),
		zap.Duration("elapsed_time", output.ElapsedTime))

	return output, nil
}

// GetIncrementalSync performs incremental sync based on last sync timestamp
func (s *syncProgressService) GetIncrementalSync(ctx context.Context, lastModified time.Time, lastID gocql.UUID, syncType string) (*SyncProgressOutput, error) {
	s.logger.Info("✨ Starting incremental sync",
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
		s.logger.Error("❌ invalid sync type", zap.String("sync_type", syncType))
		return nil, errors.NewAppError("sync_type must be 'collections' or 'files'", nil)
	}
}
