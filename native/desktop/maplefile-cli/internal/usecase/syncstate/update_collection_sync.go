// internal/usecase/syncstate/update_collection_sync.go
package syncstate

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// UpdateCollectionSyncInput represents the input for updating collection sync
type UpdateCollectionSyncInput struct {
	LastCollectionSync time.Time
	LastCollectionID   primitive.ObjectID
}

// UpdateCollectionSyncUseCase defines the interface for updating collection sync state
type UpdateCollectionSyncUseCase interface {
	Execute(ctx context.Context, input *UpdateCollectionSyncInput) error
}

// updateCollectionSyncUseCase implements the UpdateCollectionSyncUseCase interface
type updateCollectionSyncUseCase struct {
	logger              *zap.Logger
	repository          syncstate.SyncStateRepository
	getSyncStateUseCase GetSyncStateUseCase
}

// NewUpdateCollectionSyncUseCase creates a new use case for updating collection sync state
func NewUpdateCollectionSyncUseCase(
	logger *zap.Logger,
	repository syncstate.SyncStateRepository,
	getSyncStateUseCase GetSyncStateUseCase,
) UpdateCollectionSyncUseCase {
	return &updateCollectionSyncUseCase{
		logger:              logger,
		repository:          repository,
		getSyncStateUseCase: getSyncStateUseCase,
	}
}

// Execute updates the collection sync timestamp and ID
func (uc *updateCollectionSyncUseCase) Execute(ctx context.Context, input *UpdateCollectionSyncInput) error {
	// Validate input
	if input == nil {
		return errors.NewAppError("update collection sync input is required", nil)
	}

	if input.LastCollectionSync.IsZero() {
		return errors.NewAppError("last collection sync time is required", nil)
	}

	uc.logger.Debug("Updating collection sync state",
		zap.Time("lastCollectionSync", input.LastCollectionSync),
		zap.String("lastCollectionID", input.LastCollectionID.Hex()))

	// Get current sync state
	currentState, err := uc.getSyncStateUseCase.Execute(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current sync state", err)
	}

	if currentState == nil {
		currentState = &syncstate.SyncState{}
	}

	// Update only collection sync fields
	currentState.LastCollectionSync = input.LastCollectionSync
	currentState.LastCollectionID = input.LastCollectionID

	// Save updated state
	err = uc.repository.SaveSyncState(ctx, currentState)
	if err != nil {
		uc.logger.Error("Failed to update collection sync state", zap.Error(err))
		return errors.NewAppError("failed to update collection sync state", err)
	}

	uc.logger.Debug("Successfully updated collection sync state")
	return nil
}
