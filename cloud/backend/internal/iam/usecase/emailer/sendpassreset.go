package emailer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type SendFederatedUserPasswordResetEmailUseCase interface {
	Execute(ctx context.Context, monolithModule int, user *domain.FederatedUser) error
}
type sendFederatedUserPasswordResetEmailUseCaseImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	emailer templatedemailer.TemplatedEmailer
}

func NewSendFederatedUserPasswordResetEmailUseCase(config *config.Configuration, logger *zap.Logger, emailer templatedemailer.TemplatedEmailer) SendFederatedUserPasswordResetEmailUseCase {
	return &sendFederatedUserPasswordResetEmailUseCaseImpl{config, logger, emailer}
}

func (uc *sendFederatedUserPasswordResetEmailUseCaseImpl) Execute(ctx context.Context, monolithModule int, user *domain.FederatedUser) error {
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
		if user.PasswordResetVerificationCode == "" {
			e["password_reset_verification_code"] = "Password reset verification code is required"
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

	return uc.emailer.SendUserPasswordResetEmail(ctx, monolithModule, user.Email, user.PasswordResetVerificationCode, user.FirstName)
}
