// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser/impl.go
package federateduser

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

type userStorerImpl struct {
	Logger  *zap.Logger
	Session *gocql.Session
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, session *gocql.Session) dom_user.Repository {
	return &userStorerImpl{
		Logger:  loggerp,
		Session: session,
	}
}

// ListAll retrieves all users from the database
func (impl userStorerImpl) ListAll(ctx context.Context) ([]*dom_user.FederatedUser, error) {
	return nil, nil
}
