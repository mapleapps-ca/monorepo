package emailer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type SendFederatedUserVerificationEmailUseCase interface {
	Execute(ctx context.Context, monolithModule int, user *domain.FederatedUser) error
}
type sendFederatedUserVerificationEmailUseCaseImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	emailer templatedemailer.TemplatedEmailer
}

func NewSendFederatedUserVerificationEmailUseCase(config *config.Configuration, logger *zap.Logger, emailer templatedemailer.TemplatedEmailer) SendFederatedUserVerificationEmailUseCase {
	return &sendFederatedUserVerificationEmailUseCaseImpl{config, logger, emailer}
}

func (uc *sendFederatedUserVerificationEmailUseCaseImpl) Execute(ctx context.Context, monolithModule int, user *domain.FederatedUser) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if user == nil {
		e["user"] = "User is missing value"
	} else {
		if user.FirstName == "" {
			e["first_name"] = "First name is required"
		}
		if user.Email == "" {
			e["email"] = "Email is required"
		}
		if user.EmailVerificationCode == "" {
			e["email_verification_code"] = "Email verification code is required"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Send email
	//

	return uc.emailer.SendUserVerificationEmail(ctx, monolithModule, user.Email, user.EmailVerificationCode, user.FirstName)
}
