// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser/create.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FederatedUserCreateUseCase interface {
	Execute(ctx context.Context, user *dom_user.FederatedUser) error
}

type userCreateUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.FederatedUserRepository
}

func NewFederatedUserCreateUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.FederatedUserRepository) FederatedUserCreateUseCase {
	return &userCreateUseCaseImpl{config, logger, repo}
}

func (uc *userCreateUseCaseImpl) Execute(ctx context.Context, user *dom_user.FederatedUser) error {
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
