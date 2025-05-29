// native/desktop/maplefile-cli/internal/repo/sync/state.go
package syncstate

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

func (r *syncStateRepository) GetSyncState(ctx context.Context) (*syncstate.SyncState, error) {
	r.logger.Debug("💾 Getting sync state from local storage")

	// Get sync state from database
	stateBytes, err := r.dbClient.Get(syncStateKey)
	if err != nil {
		r.logger.Error("🚨 Failed to retrieve sync state from local storage", zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve sync state from local storage", err)
	}

	// If no state exists, return default empty state
	if stateBytes == nil {
		r.logger.Debug("ℹ️ No sync state found, returning default state")
		return &syncstate.SyncState{
			LastCollectionSync: time.Time{},
			LastFileSync:       time.Time{},
		}, nil
	}

	// Deserialize the sync state
	var state syncstate.SyncState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		r.logger.Error("❌ Failed to deserialize sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize sync state", err)
	}

	r.logger.Debug("✅ Successfully retrieved sync state from local storage",
		zap.Time("last_collection_sync", state.LastCollectionSync),
		zap.Time("last_file_sync", state.LastFileSync))

	return &state, nil
}
