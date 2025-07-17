// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/impl.go
package storagedailyusage

import (
	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
)

type storageDailyUsageRepositoryImpl struct {
	Logger  *zap.Logger
	Session *gocql.Session
}

func NewRepository(appCfg *config.Configuration, session *gocql.Session, loggerp *zap.Logger) storagedailyusage.StorageDailyUsageRepository {
	loggerp = loggerp.Named("StorageDailyUsageRepository")

	return &storageDailyUsageRepositoryImpl{
		Logger:  loggerp,
		Session: session,
	}
}
