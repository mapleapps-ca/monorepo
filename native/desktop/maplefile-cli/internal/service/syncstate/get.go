// internal/service/syncstate/get.go
package syncstate

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// GetOutput represents the result of getting sync state
type GetOutput struct {
	SyncState *syncstate.SyncState `json:"sync_state"`
}

// GetService defines the interface for getting sync state
type GetService interface {
	GetSyncState(ctx context.Context) (*GetOutput, error)
}

// getService implements the GetService interface
type getService struct {
	logger        *zap.Logger
	syncStateRepo syncstate.SyncStateRepository
}

// NewGetService creates a new service for getting sync state
func NewGetService(
	logger *zap.Logger,
	syncStateRepo syncstate.SyncStateRepository,
) GetService {
	logger = logger.Named("GetService")
	return &getService{
		logger:        logger,
		syncStateRepo: syncStateRepo,
	}
}

// GetSyncState retrieves the current sync state
func (s *getService) GetSyncState(ctx context.Context) (*GetOutput, error) {
	s.logger.Debug("Getting current sync state")

	// Get sync state from repository
	syncState, err := s.syncStateRepo.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("failed to get sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	s.logger.Info("Successfully retrieved sync state",
		zap.Time("last_collection_sync", syncState.LastCollectionSync),
		zap.Time("last_file_sync", syncState.LastFileSync))

	return &GetOutput{
		SyncState: syncState,
	}, nil
}
