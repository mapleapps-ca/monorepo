// native/desktop/maplefile-cli/internal/repo/recovery/impl.go
package recovery

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

// Key prefixes for different recovery entities to avoid key collisions
const (
	sessionKeyPrefix   = "recovery_session:"
	challengeKeyPrefix = "recovery_challenge:"
	tokenKeyPrefix     = "recovery_token:"
	attemptKeyPrefix   = "recovery_attempt:"
)

// recoveryRepository implements the recovery.RecoveryRepository interface for local storage
type recoveryRepository struct {
	logger   *zap.Logger
	dbClient storage.Storage
}

// NewRecoveryRepository creates a new repository for local recovery operations
func NewRecoveryRepository(
	logger *zap.Logger,
	dbClient storage.Storage,
) recovery.RecoveryRepository {
	logger = logger.Named("RecoveryRepository")
	return &recoveryRepository{
		logger:   logger,
		dbClient: dbClient,
	}
}

// Transaction support
func (r *recoveryRepository) OpenTransaction() error {
	return r.dbClient.OpenTransaction()
}

func (r *recoveryRepository) CommitTransaction() error {
	return r.dbClient.CommitTransaction()
}

func (r *recoveryRepository) DiscardTransaction() {
	r.dbClient.DiscardTransaction()
}
