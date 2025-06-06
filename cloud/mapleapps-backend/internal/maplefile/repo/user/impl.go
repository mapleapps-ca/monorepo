// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/impl.go
package user

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

type userStorerImpl struct {
	Logger *zap.Logger
	//TODO
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_user.Repository {
	loggerp = loggerp.Named("UserRepository")

	// Create the repository instance first
	repo := &userStorerImpl{
		Logger: loggerp,
		//TODO
	}
	return repo
}

// ListAll retrieves all users from the database
func (impl userStorerImpl) ListAll(ctx context.Context) ([]*dom_user.User, error) {
	return nil, nil
}
