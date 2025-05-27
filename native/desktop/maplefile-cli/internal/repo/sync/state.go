// native/desktop/maplefile-cli/internal/repo/sync/state.go
package sync

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/sync"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

const syncStateKey = "sync_state"

// syncStateRepository implements the sync.SyncStateRepository interface
type syncStateRepository struct {
	logger   *zap.Logger
	dbClient storage.Storage
}

// NewSyncStateRepository creates a new repository for sync state operations
func NewSyncStateRepository(
	logger *zap.Logger,
	dbClient storage.Storage,
) sync.SyncStateRepository {
	return &syncStateRepository{
		logger:   logger,
		dbClient: dbClient,
	}
}

func (r *syncStateRepository) GetSyncState(ctx context.Context) (*sync.SyncState, error) {
	r.logger.Debug("Getting sync state from local storage")

	// Get sync state from database
	stateBytes, err := r.dbClient.Get(syncStateKey)
	if err != nil {
		r.logger.Error("Failed to retrieve sync state from local storage", zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve sync state from local storage", err)
	}

	// If no state exists, return default empty state
	if stateBytes == nil {
		r.logger.Debug("No sync state found, returning default state")
		return &sync.SyncState{
			LastCollectionSync: time.Time{},
			LastFileSync:       time.Time{},
		}, nil
	}

	// Deserialize the sync state
	var state sync.SyncState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		r.logger.Error("Failed to deserialize sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize sync state", err)
	}

	r.logger.Debug("Successfully retrieved sync state from local storage",
		zap.Time("last_collection_sync", state.LastCollectionSync),
		zap.Time("last_file_sync", state.LastFileSync))

	return &state, nil
}

func (r *syncStateRepository) SaveSyncState(ctx context.Context, state *sync.SyncState) error {
	r.logger.Debug("Saving sync state to local storage",
		zap.Time("last_collection_sync", state.LastCollectionSync),
		zap.Time("last_file_sync", state.LastFileSync))

	// Serialize the sync state
	stateBytes, err := json.Marshal(state)
	if err != nil {
		r.logger.Error("Failed to serialize sync state", zap.Error(err))
		return errors.NewAppError("failed to serialize sync state", err)
	}

	// Save to database
	if err := r.dbClient.Set(syncStateKey, stateBytes); err != nil {
		r.logger.Error("Failed to save sync state to local storage", zap.Error(err))
		return errors.NewAppError("failed to save sync state to local storage", err)
	}

	r.logger.Debug("Successfully saved sync state to local storage")
	return nil
}

func (r *syncStateRepository) ResetSyncState(ctx context.Context) error {
	r.logger.Debug("Resetting sync state")

	// Create default empty state
	defaultState := &sync.SyncState{
		LastCollectionSync: time.Time{},
		LastFileSync:       time.Time{},
	}

	// Save the reset state
	err := r.SaveSyncState(ctx, defaultState)
	if err != nil {
		r.logger.Error("Failed to reset sync state", zap.Error(err))
		return err
	}

	r.logger.Info("Successfully reset sync state")
	return nil
}
