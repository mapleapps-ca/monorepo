// internal/service/syncdto/get_full.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// GetFullSyncInput represents the input for getting both collection and file sync data
type GetFullSyncInput struct {
	CollectionCursor   *syncdto.SyncCursorDTO `json:"collection_cursor,omitempty"`
	FileCursor         *syncdto.SyncCursorDTO `json:"file_cursor,omitempty"`
	CollectionLimit    int64                  `json:"collection_limit,omitempty"`
	FileLimit          int64                  `json:"file_limit,omitempty"`
	IncludeCollections bool                   `json:"include_collections"`
	IncludeFiles       bool                   `json:"include_files"`
}

// GetFullSyncOutput represents the result of getting both collection and file sync data
type GetFullSyncOutput struct {
	CollectionResponse *syncdto.CollectionSyncResponseDTO `json:"collection_response,omitempty"`
	FileResponse       *syncdto.FileSyncResponseDTO       `json:"file_response,omitempty"`
	CollectionsCount   int                                `json:"collections_count"`
	FilesCount         int                                `json:"files_count"`
	Message            string                             `json:"message"`
}

// GetFullSyncService defines the interface for getting comprehensive sync data
type GetFullSyncService interface {
	GetFullSyncData(ctx context.Context, input *GetFullSyncInput) (*GetFullSyncOutput, error)
	GetBothSyncData(ctx context.Context, collectionLimit, fileLimit int64) (*GetFullSyncOutput, error)
}

// getFullSyncService implements the GetFullSyncService interface
type getFullSyncService struct {
	logger      *zap.Logger
	syncDTORepo syncdto.SyncDTORepository
}

// NewGetFullSyncService creates a new service for getting comprehensive sync data
func NewGetFullSyncService(
	logger *zap.Logger,
	syncDTORepo syncdto.SyncDTORepository,
) GetFullSyncService {
	logger = logger.Named("GetFullSyncService")
	return &getFullSyncService{
		logger:      logger,
		syncDTORepo: syncDTORepo,
	}
}

// GetFullSyncData retrieves both collection and file sync data based on input parameters
func (s *getFullSyncService) GetFullSyncData(ctx context.Context, input *GetFullSyncInput) (*GetFullSyncOutput, error) {
	// Set default values
	if input == nil {
		input = &GetFullSyncInput{
			IncludeCollections: true,
			IncludeFiles:       true,
		}
	}

	if !input.IncludeCollections && !input.IncludeFiles {
		s.logger.Error("❌ at least one sync type must be included")
		return nil, errors.NewAppError("at least one of include_collections or include_files must be true", nil)
	}

	if input.CollectionLimit <= 0 {
		input.CollectionLimit = 100 // Default limit
	}
	if input.FileLimit <= 0 {
		input.FileLimit = 100 // Default limit
	}

	s.logger.Info("🔄 Starting full sync data retrieval",
		zap.Bool("include_collections", input.IncludeCollections),
		zap.Bool("include_files", input.IncludeFiles),
		zap.Int64("collection_limit", input.CollectionLimit),
		zap.Int64("file_limit", input.FileLimit))

	output := &GetFullSyncOutput{}

	// Get collection sync data if requested
	if input.IncludeCollections {
		s.logger.Debug("📁 Getting collection sync data")
		collectionResponse, err := s.syncDTORepo.GetCollectionSyncDataFromCloud(ctx, input.CollectionCursor, input.CollectionLimit)
		if err != nil {
			s.logger.Error("❌ failed to get collection sync data", zap.Error(err))
			return nil, errors.NewAppError("failed to get collection sync data", err)
		}
		output.CollectionResponse = collectionResponse
		output.CollectionsCount = len(collectionResponse.Collections)
	}

	// Get file sync data if requested
	if input.IncludeFiles {
		s.logger.Debug("📄 Getting file sync data")
		fileResponse, err := s.syncDTORepo.GetFileSyncDataFromCloud(ctx, input.FileCursor, input.FileLimit)
		if err != nil {
			s.logger.Error("❌ failed to get file sync data", zap.Error(err))
			return nil, errors.NewAppError("failed to get file sync data", err)
		}
		output.FileResponse = fileResponse
		output.FilesCount = len(fileResponse.Files)
	}

	// Generate summary message
	totalItems := output.CollectionsCount + output.FilesCount
	if totalItems == 0 {
		output.Message = "No changes found since last sync"
	} else {
		output.Message = "Full sync data retrieved successfully"
	}

	s.logger.Info("✅ Successfully completed full sync data retrieval",
		zap.Int("collections_count", output.CollectionsCount),
		zap.Int("files_count", output.FilesCount),
		zap.Int("total_items", totalItems))

	return output, nil
}

// GetBothSyncData is a convenience method to get both collections and files with default settings
func (s *getFullSyncService) GetBothSyncData(ctx context.Context, collectionLimit, fileLimit int64) (*GetFullSyncOutput, error) {
	s.logger.Debug("🔄 Getting both collection and file sync data with default settings",
		zap.Int64("collection_limit", collectionLimit),
		zap.Int64("file_limit", fileLimit))

	input := &GetFullSyncInput{
		IncludeCollections: true,
		IncludeFiles:       true,
		CollectionLimit:    collectionLimit,
		FileLimit:          fileLimit,
	}

	return s.GetFullSyncData(ctx, input)
}
