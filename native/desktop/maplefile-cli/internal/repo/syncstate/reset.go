// native/desktop/maplefile-cli/internal/repo/sync/state.go
package syncstate

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

func (r *syncStateRepository) ResetSyncState(ctx context.Context) error {
	r.logger.Debug("🔄 Resetting sync state")

	// Create default empty state
	defaultState := &syncstate.SyncState{
		LastCollectionSync: time.Time{},
		LastFileSync:       time.Time{},
	}

	// Save the reset state
	err := r.SaveSyncState(ctx, defaultState)
	if err != nil {
		r.logger.Error("❌ Failed to reset sync state", zap.Error(err))
		return err
	}

	r.logger.Info("✅ Successfully reset sync state")
	return nil
}
