// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storageusageevent/impl.go
package storageusageevent

import (
	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent"
)

type storageUsageEventRepositoryImpl struct {
	Logger  *zap.Logger
	Session *gocql.Session
}

func NewRepository(appCfg *config.Configuration, session *gocql.Session, loggerp *zap.Logger) storageusageevent.StorageUsageEventRepository {
	loggerp = loggerp.Named("StorageUsageEventRepository")

	return &storageUsageEventRepositoryImpl{
		Logger:  loggerp,
		Session: session,
	}
}
