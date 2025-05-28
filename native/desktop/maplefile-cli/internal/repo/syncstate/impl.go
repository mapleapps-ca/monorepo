// native/desktop/maplefile-cli/internal/repo/sync/state.go
package syncstate

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

const syncStateKey = "sync_state"

// syncStateRepository implements the syncstate.SyncStateRepository interface
type syncStateRepository struct {
	logger   *zap.Logger
	dbClient storage.Storage
}

// NewSyncStateRepository creates a new repository for sync state operations
func NewSyncStateRepository(
	logger *zap.Logger,
	dbClient storage.Storage,
) syncstate.SyncStateRepository {
	logger = logger.Named("SyncStateRepository")
	return &syncStateRepository{
		logger:   logger,
		dbClient: dbClient,
	}
}
