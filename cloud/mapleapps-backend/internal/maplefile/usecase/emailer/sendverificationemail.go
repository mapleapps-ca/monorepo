package emailer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type SendUserVerificationEmailUseCase interface {
	Execute(ctx context.Context, user *domain.User) error
}
type sendUserVerificationEmailUseCaseImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	emailer templatedemailer.TemplatedEmailer
}

func NewSendUserVerificationEmailUseCase(config *config.Configuration, logger *zap.Logger, emailer templatedemailer.TemplatedEmailer) SendUserVerificationEmailUseCase {
	logger = logger.Named("SendUserVerificationEmailUseCase")
	return &sendUserVerificationEmailUseCaseImpl{config, logger, emailer}
}

func (uc *sendUserVerificationEmailUseCaseImpl) Execute(ctx context.Context, user *domain.User) error {
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
		if user.SecurityData.Code == "" {
			e["code"] = "Code is required for password reset verification "
		}
		if user.SecurityData.CodeType != domain.UserCodeTypePasswordReset {
			e["code_type"] = "Code type is required for password reset verification "
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

	return uc.emailer.SendUserVerificationEmail(ctx, user.Email, user.SecurityData.Code, user.FirstName)
}
