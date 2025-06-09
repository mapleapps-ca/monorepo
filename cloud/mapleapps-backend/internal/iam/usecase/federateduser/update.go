package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FederatedUserUpdateUseCase interface {
	Execute(ctx context.Context, user *dom_user.FederatedUser) error
}

type userUpdateUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.FederatedUserRepository
}

func NewFederatedUserUpdateUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.FederatedUserRepository) FederatedUserUpdateUseCase {
	logger = logger.Named("FederatedUserUpdateUseCase")
	return &userUpdateUseCaseImpl{config, logger, repo}
}

func (uc *userUpdateUseCaseImpl) Execute(ctx context.Context, user *dom_user.FederatedUser) error {
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
