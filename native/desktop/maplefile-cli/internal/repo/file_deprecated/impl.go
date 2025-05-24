// monorepo/native/desktop/maplefile-cli/internal/repo/file/impl.go
package file

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

// Key prefix for files to avoid key collisions with other types
const fileKeyPrefix = "local_file:"

// fileRepository implements the file.FileRepository interface
type fileRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	dbClient      storage.Storage // LevelDB client for metadata storage
}

// NewFileRepository creates a new repository for file operations
func NewFileRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	dbClient storage.Storage,
) file.FileRepository {
	return &fileRepository{
		logger:        logger,
		configService: configService,
		dbClient:      dbClient,
	}
}

func (r *fileRepository) OpenTransaction() error {
	return r.dbClient.OpenTransaction()
}

func (r *fileRepository) CommitTransaction() error {
	return r.dbClient.CommitTransaction()
}

func (r *fileRepository) DiscardTransaction() {
	r.dbClient.DiscardTransaction()
}
