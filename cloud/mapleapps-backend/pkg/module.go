// Package pkg provides core infrastructure components and dependencies used across the backend services.
package pkg

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/distributedmutex"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/emailer/mailgun"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/blacklist"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/ipcountryblocker"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/twotiercache"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/database/cassandradb"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/memory/redis"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/object/s3"
)

func Module() fx.Option {
	return fx.Options(
		// Emailer providers with proper naming
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

		// Core infrastructure providers
		fx.Provide(
			// Security components
			blacklist.NewProvider,
			distributedmutex.NewAdapter,
			ipcountryblocker.NewProvider,
			jwt.NewJWTProvider,
			password.NewPasswordProvider,

			// Database components
			cassandradb.NewCassandraConnection,
			cassandradb.NewMigrator,

			// Cache and storage components
			redis.NewCache,
			twotiercache.NewTwoTierCache,
			cassandracache.NewCassandraCacher,
			s3.NewProvider,

			// Observability components (depends on infrastructure for health checks)
			observability.NewHealthChecker,
			observability.NewMetricsServer,
		),

		// Lifecycle management for infrastructure components
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
			// Run migrations in background
			go func() {
				if err := migrator.Up(); err != nil {
					logger.Error("Migration failed", zap.Error(err))
				}
			}()
			return nil // Return immediately
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down infrastructure components...")
			if session != nil {
				session.Close()
			}
			logger.Info("Infrastructure components shutdown complete")
			return nil
		},
	})
}
