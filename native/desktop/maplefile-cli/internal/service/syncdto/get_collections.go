// internal/service/syncdto/get_collections.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// GetCollectionsInput represents the input for getting collection sync data
type GetCollectionsInput struct {
	Cursor *syncdto.SyncCursorDTO `json:"cursor,omitempty"`
	Limit  int64                  `json:"limit,omitempty"`
}

// GetCollectionsOutput represents the result of getting collection sync data
type GetCollectionsOutput struct {
	Response *syncdto.CollectionSyncResponseDTO `json:"response"`
	Message  string                             `json:"message"`
}

// GetCollectionsService defines the interface for getting collection sync data
type GetCollectionsService interface {
	GetCollectionSyncData(ctx context.Context, input *GetCollectionsInput) (*GetCollectionsOutput, error)
}

// getCollectionsService implements the GetCollectionsService interface
type getCollectionsService struct {
	logger      *zap.Logger
	syncDTORepo syncdto.SyncDTORepository
}

// NewGetCollectionsService creates a new service for getting collection sync data
func NewGetCollectionsService(
	logger *zap.Logger,
	syncDTORepo syncdto.SyncDTORepository,
) GetCollectionsService {
	return &getCollectionsService{
		logger:      logger,
		syncDTORepo: syncDTORepo,
	}
}

// GetCollectionSyncData retrieves collection sync data from the cloud
func (s *getCollectionsService) GetCollectionSyncData(ctx context.Context, input *GetCollectionsInput) (*GetCollectionsOutput, error) {
	// Set default values
	if input == nil {
		input = &GetCollectionsInput{}
	}

	if input.Limit <= 0 {
		input.Limit = 100 // Default limit
	}

	s.logger.Debug("Getting collection sync data from cloud",
		zap.Any("cursor", input.Cursor),
		zap.Int64("limit", input.Limit))

	// Get collection sync data from repository
	response, err := s.syncDTORepo.GetCollectionSyncDataFromCloud(ctx, input.Cursor, input.Limit)
	if err != nil {
		s.logger.Error("failed to get collection sync data from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get collection sync data from cloud", err)
	}

	if response == nil {
		s.logger.Warn("received nil response from collection sync data")
		return nil, errors.NewAppError("received empty response from cloud", nil)
	}

	s.logger.Info("Successfully retrieved collection sync data",
		zap.Int("collections_count", len(response.Collections)),
		zap.Bool("has_more", response.HasMore))

	message := "Collection sync data retrieved successfully"
	if len(response.Collections) == 0 {
		message = "No collection changes found since last sync"
	}

	return &GetCollectionsOutput{
		Response: response,
		Message:  message,
	}, nil
}
