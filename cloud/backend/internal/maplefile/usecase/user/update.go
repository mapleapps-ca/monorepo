package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UserUpdateUseCase interface {
	Execute(ctx context.Context, user *dom_user.User) error
}

type userUpdateUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserUpdateUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) UserUpdateUseCase {
	logger = logger.Named("UserUpdateUseCase")
	return &userUpdateUseCaseImpl{config, logger, repo}
}

func (uc *userUpdateUseCaseImpl) Execute(ctx context.Context, user *dom_user.User) error {
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
	// STEP 2: Update in database.
	//

	return uc.repo.UpdateByID(ctx, user)
}
