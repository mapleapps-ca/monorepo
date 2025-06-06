// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user/getbyid.go
package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UserDeleteUserByEmailUseCase interface {
	Execute(ctx context.Context, email string) error
}

type userDeleteUserByEmailImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserDeleteUserByEmailUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) UserDeleteUserByEmailUseCase {
	logger = logger.Named("UserDeleteUserByEmailUseCase")
	return &userDeleteUserByEmailImpl{config, logger, repo}
}

func (uc *userDeleteUserByEmailImpl) Execute(ctx context.Context, email string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if email == "" {
		e["email"] = "missing value"
	} else {
		//TODO: IMPL.
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.DeleteByEmail(ctx, email)
}
