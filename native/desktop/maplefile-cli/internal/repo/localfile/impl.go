// monorepo/native/desktop/maplefile-cli/internal/repo/localfile/impl.go
package localfile

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

// Key prefix for files to avoid key collisions with other types
const fileKeyPrefix = "local_file:"

// localFileRepository implements the localfile.LocalFileRepository interface
type localFileRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	dbClient      storage.Storage // LevelDB client for metadata storage
}

// NewLocalFileRepository creates a new repository for localfile operations
func NewLocalFileRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	dbClient storage.Storage,
) localfile.LocalFileRepository {
	return &localFileRepository{
		logger:        logger,
		configService: configService,
		dbClient:      dbClient,
	}
}

func (r *localFileRepository) OpenTransaction() error {
	return r.dbClient.OpenTransaction()
}

func (r *localFileRepository) CommitTransaction() error {
	return r.dbClient.CommitTransaction()
}

func (r *localFileRepository) DiscardTransaction() {
	r.dbClient.DiscardTransaction()
}
