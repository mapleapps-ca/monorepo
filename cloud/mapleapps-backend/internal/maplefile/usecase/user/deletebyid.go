// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user/getbyid.go
package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UserDeleteByIDUseCase interface {
	Execute(ctx context.Context, id gocql.UUID) error
}

type userDeleteByIDImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserDeleteByIDUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) UserDeleteByIDUseCase {
	logger = logger.Named("UserDeleteByIDUseCase")
	return &userDeleteByIDImpl{config, logger, repo}
}

func (uc *userDeleteByIDImpl) Execute(ctx context.Context, id gocql.UUID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "missing value"
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.DeleteByID(ctx, id)
}
