// monorepo/native/desktop/maplefile-cli/internal/repo/collection/impl.go
package collection

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

// Key prefix for collections to avoid key collisions with other types
const collectionKeyPrefix = "local_collection:"

// collectionRepository implements the collection.CollectionRepository interface
type collectionRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
	dbClient      storage.Storage // Add LevelDB client for local storage
}

// NewCollectionRepository creates a new repository for collection operations
func NewCollectionRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	dbClient storage.Storage, // Add storage client parameter
) collection.CollectionRepository {
	logger = logger.Named("CollectionRepository")
	return &collectionRepository{
		logger:        logger,
		configService: configService,
		dbClient:      dbClient,
	}
}

func (r *collectionRepository) OpenTransaction() error {
	return r.dbClient.OpenTransaction()
}

func (r *collectionRepository) CommitTransaction() error {
	return r.dbClient.CommitTransaction()
}

func (r *collectionRepository) DiscardTransaction() {
	r.dbClient.DiscardTransaction()
}
