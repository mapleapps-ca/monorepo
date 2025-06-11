// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/impl.go
package filemetadata

import (
	"encoding/json"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

type fileMetadataRepositoryImpl struct {
	Logger  *zap.Logger
	Session *gocql.Session
}

func NewRepository(appCfg *config.Configuration, session *gocql.Session, loggerp *zap.Logger) dom_file.FileMetadataRepository {
	loggerp = loggerp.Named("FileMetadataRepository")

	return &fileMetadataRepositoryImpl{
		Logger:  loggerp,
		Session: session,
	}
}

// Helper functions for JSON serialization
func (impl *fileMetadataRepositoryImpl) serializeEncryptedFileKey(key keys.EncryptedFileKey) (string, error) {
	data, err := json.Marshal(key)
	return string(data), err
}

func (impl *fileMetadataRepositoryImpl) deserializeEncryptedFileKey(data string) (keys.EncryptedFileKey, error) {
	if data == "" {
		return keys.EncryptedFileKey{}, nil
	}
	var key keys.EncryptedFileKey
	err := json.Unmarshal([]byte(data), &key)
	return key, err
}

// isValidUUID checks if UUID is not nil/empty
func (impl *fileMetadataRepositoryImpl) isValidUUID(id gocql.UUID) bool {
	return id.String() != "00000000-0000-0000-0000-000000000000"
}
