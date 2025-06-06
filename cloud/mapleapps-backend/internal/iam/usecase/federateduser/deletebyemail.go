// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser/getbyid.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FederatedUserDeleteFederatedUserByEmailUseCase interface {
	Execute(ctx context.Context, email string) error
}

type userDeleteFederatedUserByEmailImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewFederatedUserDeleteFederatedUserByEmailUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) FederatedUserDeleteFederatedUserByEmailUseCase {
	return &userDeleteFederatedUserByEmailImpl{config, logger, repo}
}

func (uc *userDeleteFederatedUserByEmailImpl) Execute(ctx context.Context, email string) error {
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
