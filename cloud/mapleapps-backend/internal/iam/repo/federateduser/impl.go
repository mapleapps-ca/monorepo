// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser/impl.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

type userStorerImpl struct {
	Logger *zap.Logger
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_user.Repository {

	return &userStorerImpl{
		Logger: loggerp,
	}
}

// ListAll retrieves all users from the database
func (impl userStorerImpl) ListAll(ctx context.Context) ([]*dom_user.FederatedUser, error) {
	return nil, nil
}
