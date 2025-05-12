// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser/getbyemail.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FederatedUserGetByEmailUseCase interface {
	Execute(ctx context.Context, email string) (*dom_user.FederatedUser, error)
}

type userGetByEmailUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewFederatedUserGetByEmailUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) FederatedUserGetByEmailUseCase {
	return &userGetByEmailUseCaseImpl{config, logger, repo}
}

func (uc *userGetByEmailUseCaseImpl) Execute(ctx context.Context, email string) (*dom_user.FederatedUser, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if email == "" {
		e["email"] = "missing value"
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByEmail(ctx, email)
}
