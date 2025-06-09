// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/impl.go
package user

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

type Params struct {
	fx.In
	Session *gocql.Session
	Logger  *zap.Logger
}

type userStorerImpl struct {
	session *gocql.Session
	logger  *zap.Logger
}

func NewRepository(p Params) dom_user.Repository {
	p.Logger = p.Logger.Named("MapleFileUserRepository")
	return &userStorerImpl{
		session: p.Session,
		logger:  p.Logger,
	}
}

// ListAll retrieves all users from the database
func (impl userStorerImpl) ListAll(ctx context.Context) ([]*dom_user.User, error) {
	return nil, nil
}
