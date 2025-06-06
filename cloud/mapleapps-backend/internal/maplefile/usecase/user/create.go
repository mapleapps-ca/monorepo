// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user/create.go
package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UserCreateUseCase interface {
	Execute(ctx context.Context, user *dom_user.User) error
}

type userCreateUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserCreateUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) UserCreateUseCase {
	logger = logger.Named("UserCreateUseCase")
	return &userCreateUseCaseImpl{config, logger, repo}
}

func (uc *userCreateUseCaseImpl) Execute(ctx context.Context, user *dom_user.User) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if user == nil {
		e["user"] = "missing value"
	} else {
		//TODO: IMPL.
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}
	//
	// STEP 2: Insert into database.
	//

	return uc.repo.Create(ctx, user)
}
