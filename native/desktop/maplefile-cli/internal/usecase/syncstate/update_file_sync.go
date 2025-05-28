// internal/usecase/syncstate/update_file_sync.go
package syncstate

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// UpdateFileSyncInput represents the input for updating file sync
type UpdateFileSyncInput struct {
	LastFileSync time.Time
	LastFileID   primitive.ObjectID
}

// UpdateFileSyncUseCase defines the interface for updating file sync state
type UpdateFileSyncUseCase interface {
	Execute(ctx context.Context, input *UpdateFileSyncInput) error
}

// updateFileSyncUseCase implements the UpdateFileSyncUseCase interface
type updateFileSyncUseCase struct {
	logger              *zap.Logger
	repository          syncstate.SyncStateRepository
	getSyncStateUseCase GetSyncStateUseCase
}

// NewUpdateFileSyncUseCase creates a new use case for updating file sync state
func NewUpdateFileSyncUseCase(
	logger *zap.Logger,
	repository syncstate.SyncStateRepository,
	getSyncStateUseCase GetSyncStateUseCase,
) UpdateFileSyncUseCase {
	logger = logger.Named("UpdateFileSyncUseCase")
	return &updateFileSyncUseCase{
		logger:              logger,
		repository:          repository,
		getSyncStateUseCase: getSyncStateUseCase,
	}
}

// Execute updates the file sync timestamp and ID
func (uc *updateFileSyncUseCase) Execute(ctx context.Context, input *UpdateFileSyncInput) error {
	// Validate input
	if input == nil {
		return errors.NewAppError("update file sync input is required", nil)
	}

	if input.LastFileSync.IsZero() {
		return errors.NewAppError("last file sync time is required", nil)
	}

	uc.logger.Debug("Updating file sync state",
		zap.Time("lastFileSync", input.LastFileSync),
		zap.String("lastFileID", input.LastFileID.Hex()))

	// Get current sync state
	currentState, err := uc.getSyncStateUseCase.Execute(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current sync state", err)
	}

	if currentState == nil {
		currentState = &syncstate.SyncState{}
	}

	// Update only file sync fields
	currentState.LastFileSync = input.LastFileSync
	currentState.LastFileID = input.LastFileID

	// Save updated state
	err = uc.repository.SaveSyncState(ctx, currentState)
	if err != nil {
		uc.logger.Error("Failed to update file sync state", zap.Error(err))
		return errors.NewAppError("failed to update file sync state", err)
	}

	uc.logger.Debug("Successfully updated file sync state")
	return nil
}
