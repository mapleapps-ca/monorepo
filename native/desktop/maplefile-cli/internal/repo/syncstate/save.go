// native/desktop/maplefile-cli/internal/repo/sync/state.go
package syncstate

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

func (r *syncStateRepository) SaveSyncState(ctx context.Context, state *syncstate.SyncState) error {
	r.logger.Debug("💾 Saving sync state to local storage",
		zap.Time("last_collection_sync", state.LastCollectionSync),
		zap.Time("last_file_sync", state.LastFileSync))

	// Serialize the sync state
	stateBytes, err := json.Marshal(state)
	if err != nil {
		r.logger.Error("❌ Failed to serialize sync state", zap.Error(err))
		return errors.NewAppError("failed to serialize sync state", err)
	}

	// Save to database
	if err := r.dbClient.Set(syncStateKey, stateBytes); err != nil {
		r.logger.Error("❌ Failed to save sync state to local storage", zap.Error(err))
		return errors.NewAppError("failed to save sync state to local storage", err)
	}

	r.logger.Debug("✅ Successfully saved sync state to local storage")
	return nil
}
