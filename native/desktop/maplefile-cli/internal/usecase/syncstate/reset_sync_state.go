// internal/usecase/syncstate/reset_sync_state.go
package syncstate

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// ResetSyncStateUseCase defines the interface for resetting sync state
type ResetSyncStateUseCase interface {
	Execute(ctx context.Context) error
}

// resetSyncStateUseCase implements the ResetSyncStateUseCase interface
type resetSyncStateUseCase struct {
	logger     *zap.Logger
	repository syncstate.SyncStateRepository
}

// NewResetSyncStateUseCase creates a new use case for resetting sync state
func NewResetSyncStateUseCase(
	logger *zap.Logger,
	repository syncstate.SyncStateRepository,
) ResetSyncStateUseCase {
	logger = logger.Named("ResetSyncStateUseCase")
	return &resetSyncStateUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute resets the sync state to empty/default values
func (uc *resetSyncStateUseCase) Execute(ctx context.Context) error {
	uc.logger.Debug("Resetting sync state")

	// Reset sync state through repository
	err := uc.repository.ResetSyncState(ctx)
	if err != nil {
		uc.logger.Error("Failed to reset sync state", zap.Error(err))
		return errors.NewAppError("failed to reset sync state", err)
	}

	uc.logger.Info("Successfully reset sync state")
	return nil
}
