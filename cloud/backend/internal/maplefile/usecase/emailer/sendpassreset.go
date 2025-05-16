package emailer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type SendUserPasswordResetEmailUseCase interface {
	Execute(ctx context.Context, user *domain.User) error
}
type sendUserPasswordResetEmailUseCaseImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	emailer templatedemailer.TemplatedEmailer
}

func NewSendUserPasswordResetEmailUseCase(config *config.Configuration, logger *zap.Logger, emailer templatedemailer.TemplatedEmailer) SendUserPasswordResetEmailUseCase {
	return &sendUserPasswordResetEmailUseCaseImpl{config, logger, emailer}
}

func (uc *sendUserPasswordResetEmailUseCaseImpl) Execute(ctx context.Context, user *domain.User) error {
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

	return uc.emailer.SendUserPasswordResetEmail(ctx, user.Email, user.PasswordResetVerificationCode, user.FirstName)
}
