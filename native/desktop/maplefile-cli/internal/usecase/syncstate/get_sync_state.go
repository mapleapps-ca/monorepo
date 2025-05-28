// internal/usecase/syncstate/get_sync_state.go
package syncstate

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// GetSyncStateUseCase defines the interface for getting sync state
type GetSyncStateUseCase interface {
	Execute(ctx context.Context) (*syncstate.SyncState, error)
}

// getSyncStateUseCase implements the GetSyncStateUseCase interface
type getSyncStateUseCase struct {
	logger     *zap.Logger
	repository syncstate.SyncStateRepository
}

// NewGetSyncStateUseCase creates a new use case for getting sync state
func NewGetSyncStateUseCase(
	logger *zap.Logger,
	repository syncstate.SyncStateRepository,
) GetSyncStateUseCase {
	logger = logger.Named("GetSyncStateUseCase")
	return &getSyncStateUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves the current sync state
func (uc *getSyncStateUseCase) Execute(ctx context.Context) (*syncstate.SyncState, error) {
	uc.logger.Debug("Getting sync state")

	// Get sync state from repository
	state, err := uc.repository.GetSyncState(ctx)
	if err != nil {
		uc.logger.Error("Failed to get sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	if state == nil {
		uc.logger.Debug("No sync state found, returning empty state")
		return &syncstate.SyncState{}, nil
	}

	uc.logger.Debug("Successfully retrieved sync state",
		zap.Time("lastCollectionSync", state.LastCollectionSync),
		zap.Time("lastFileSync", state.LastFileSync))

	return state, nil
}
