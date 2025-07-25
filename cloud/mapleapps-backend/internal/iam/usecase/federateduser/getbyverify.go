package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FederatedUserGetByVerificationCodeUseCase interface {
	Execute(ctx context.Context, verificationCode string) (*dom_user.FederatedUser, error)
}

type userGetByVerificationCodeUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.FederatedUserRepository
}

func NewFederatedUserGetByVerificationCodeUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.FederatedUserRepository) FederatedUserGetByVerificationCodeUseCase {
	logger = logger.Named("FederatedUserGetByVerificationCodeUseCase")
	return &userGetByVerificationCodeUseCaseImpl{config, logger, repo}
}

func (uc *userGetByVerificationCodeUseCaseImpl) Execute(ctx context.Context, verificationCode string) (*dom_user.FederatedUser, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if verificationCode == "" {
		e["verification_code"] = "missing value"
	} else {
		//TODO: IMPL.
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for get by verification",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 3: Get from database.
	//

	return uc.repo.GetByVerificationCode(ctx, verificationCode)
}
