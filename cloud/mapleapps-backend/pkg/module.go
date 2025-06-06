// Package pkg provides core infrastructure components and dependencies used across the backend services.
package pkg

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/distributedmutex"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/emailer/mailgun"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/blacklist"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/ipcountryblocker"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/database/cassandradb"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/object/s3"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				mailgun.NewMapleFileModuleEmailer,
				fx.ResultTags(`name:"maplefile-module-emailer"`),
			),
			fx.Annotate(
				mailgun.NewPaperCloudModuleEmailer,
				fx.ResultTags(`name:"papercloud-module-emailer"`),
			),
		),
		fx.Provide(
			blacklist.NewProvider,
			distributedmutex.NewAdapter,
			ipcountryblocker.NewProvider,
			jwt.NewProvider,
			password.NewProvider,
			cassandradb.NewCassandraConnection,
			cassandradb.NewMigrator,
			cassandracache.NewCache,
			s3.NewProvider,
		),
		// Add lifecycle management for Cassandra
		fx.Invoke(runMigrationsAndSetupLifecycle),
	)
}

// runMigrationsAndSetupLifecycle handles both migration execution and proper shutdown
func runMigrationsAndSetupLifecycle(
	lc fx.Lifecycle,
	session *gocql.Session,
	migrator *cassandradb.Migrator,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Running database migrations...")
			if err := migrator.Up(); err != nil {
				logger.Error("Failed to run migrations", zap.Error(err))
				return err
			}
			logger.Info("Database migrations completed successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down Cassandra connection...")
			if session != nil {
				session.Close()
			}
			logger.Info("Cassandra connection closed")
			return nil
		},
	})
}
