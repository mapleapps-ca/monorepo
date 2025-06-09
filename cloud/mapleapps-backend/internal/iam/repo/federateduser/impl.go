// repo/federateduser/impl.go
package federateduser

import (
	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

type Params struct {
	fx.In
	Session *gocql.Session
	Logger  *zap.Logger
}

type federatedUserRepository struct {
	session *gocql.Session
	logger  *zap.Logger
}

// NewRepository creates a new Cassandra repository for federated users
func NewRepository(p Params) dom.FederatedUserRepository {
	p.Logger = p.Logger.Named("FederatedUserRepository")
	return &federatedUserRepository{
		session: p.Session,
		logger:  p.Logger,
	}
}

// prepareStatements prepares commonly used statements for better performance
func (r *federatedUserRepository) prepareStatements() error {
	// This is optional but can improve performance for frequently used queries
	// Cassandra automatically caches prepared statements, but explicit preparation
	// can help with startup validation

	testQueries := []string{
		"SELECT id FROM users_by_id WHERE id = ? LIMIT 1",
		"SELECT id FROM users_by_email WHERE email = ? LIMIT 1",
	}

	for _, query := range testQueries {
		if err := r.session.Query(query, gocql.TimeUUID()).Exec(); err != nil {
			// Ignore not found errors, we're just testing the query compiles
			if err != gocql.ErrNotFound {
				r.logger.Warn("Query preparation test failed",
					zap.String("query", query),
					zap.Error(err))
			}
		}
	}

	return nil
}
