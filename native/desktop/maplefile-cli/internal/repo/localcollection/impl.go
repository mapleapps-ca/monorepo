// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/impl.go
package collection

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

// Key prefix for collections to avoid key collisions with other types
const collectionKeyPrefix = "local_collection:"

// localcollectionRepository implements the localcollection.LocalCollectionRepository interface
type localcollectionRepository struct {
	logger         *zap.Logger
	configService  config.ConfigService
	userRepo       user.Repository
	tokenRefresher auth.TokenRefresher
	httpClient     *http.Client
	dbClient       storage.Storage // Add LevelDB client for local storage
}

// NewLocalCollectionRepository creates a new repository for localcollection operations
func NewLocalCollectionRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefresher auth.TokenRefresher,
	dbClient storage.Storage, // Add storage client parameter
) localcollection.LocalCollectionRepository {
	return &localcollectionRepository{
		logger:         logger,
		configService:  configService,
		userRepo:       userRepo,
		tokenRefresher: tokenRefresher,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		dbClient:       dbClient,
	}
}

// generateKey creates a storage key for a collection
func (r *localcollectionRepository) generateKey(id string) string {
	return fmt.Sprintf("%s%s", collectionKeyPrefix, id)
}
