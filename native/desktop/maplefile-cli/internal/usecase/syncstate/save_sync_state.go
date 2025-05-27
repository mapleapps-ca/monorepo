// internal/usecase/syncstate/save_sync_state.go
package syncstate

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// SaveSyncStateUseCase defines the interface for saving sync state
type SaveSyncStateUseCase interface {
	Execute(ctx context.Context, state *syncstate.SyncState) error
}

// saveSyncStateUseCase implements the SaveSyncStateUseCase interface
type saveSyncStateUseCase struct {
	logger     *zap.Logger
	repository syncstate.SyncStateRepository
}

// NewSaveSyncStateUseCase creates a new use case for saving sync state
func NewSaveSyncStateUseCase(
	logger *zap.Logger,
	repository syncstate.SyncStateRepository,
) SaveSyncStateUseCase {
	return &saveSyncStateUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute saves the sync state
func (uc *saveSyncStateUseCase) Execute(ctx context.Context, state *syncstate.SyncState) error {
	// Validate input
	if state == nil {
		return errors.NewAppError("sync state is required", nil)
	}

	uc.logger.Debug("Saving sync state",
		zap.Time("lastCollectionSync", state.LastCollectionSync),
		zap.Time("lastFileSync", state.LastFileSync))

	// Save sync state through repository
	err := uc.repository.SaveSyncState(ctx, state)
	if err != nil {
		uc.logger.Error("Failed to save sync state", zap.Error(err))
		return errors.NewAppError("failed to save sync state", err)
	}

	uc.logger.Debug("Successfully saved sync state")
	return nil
}
