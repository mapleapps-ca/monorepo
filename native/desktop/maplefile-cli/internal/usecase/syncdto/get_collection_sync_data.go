// internal/usecase/syncdto/get_collection_sync_data.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// GetCollectionSyncDataInput represents the input for getting collection sync data
type GetCollectionSyncDataInput struct {
	Cursor *syncdto.SyncCursorDTO
	Limit  int64
}

// GetCollectionSyncDataUseCase defines the interface for getting collection sync data from cloud
type GetCollectionSyncDataUseCase interface {
	Execute(ctx context.Context, input *GetCollectionSyncDataInput) (*syncdto.CollectionSyncResponseDTO, error)
}

// getCollectionSyncDataUseCase implements the GetCollectionSyncDataUseCase interface
type getCollectionSyncDataUseCase struct {
	logger     *zap.Logger
	repository syncdto.SyncDTORepository
}

// NewGetCollectionSyncDataUseCase creates a new use case for getting collection sync data
func NewGetCollectionSyncDataUseCase(
	logger *zap.Logger,
	repository syncdto.SyncDTORepository,
) GetCollectionSyncDataUseCase {
	logger = logger.Named("GetCollectionSyncDataUseCase")
	return &getCollectionSyncDataUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves collection sync data from the cloud
func (uc *getCollectionSyncDataUseCase) Execute(ctx context.Context, input *GetCollectionSyncDataInput) (*syncdto.CollectionSyncResponseDTO, error) {
	// Validate input
	if input == nil {
		return nil, errors.NewAppError("get collection sync data input is required", nil)
	}

	// Set default limit if not provided
	limit := input.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	// Validate limit is reasonable
	if limit > 1000 {
		return nil, errors.NewAppError("limit cannot exceed 1000", nil)
	}

	uc.logger.Debug("Getting collection sync data from cloud",
		zap.Any("cursor", input.Cursor),
		zap.Int64("limit", limit))

	// Get collection sync data from repository
	response, err := uc.repository.GetCollectionSyncDataFromCloud(ctx, input.Cursor, limit)
	if err != nil {
		uc.logger.Error("Failed to get collection sync data from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get collection sync data from cloud", err)
	}

	if response == nil {
		uc.logger.Warn("Received nil response from collection sync data repository")
		return &syncdto.CollectionSyncResponseDTO{
			Collections: []syncdto.CollectionSyncItem{},
			HasMore:     false,
		}, nil
	}

	uc.logger.Debug("Successfully retrieved collection sync data",
		zap.Int("collectionsCount", len(response.Collections)),
		zap.Bool("hasMore", response.HasMore))

	return response, nil
}
