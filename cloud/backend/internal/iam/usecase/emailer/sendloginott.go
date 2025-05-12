// cloud/backend/internal/iam/usecase/emailer/sendloginott.go
package emailer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type SendLoginOTTEmailUseCase interface {
	Execute(ctx context.Context, monolithModule int, email, oneTimeToken, firstName string) error
}

type sendLoginOTTEmailUseCaseImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	emailer templatedemailer.TemplatedEmailer
}

func NewSendLoginOTTEmailUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	emailer templatedemailer.TemplatedEmailer,
) SendLoginOTTEmailUseCase {
	return &sendLoginOTTEmailUseCaseImpl{config, logger, emailer}
}

func (uc *sendLoginOTTEmailUseCaseImpl) Execute(ctx context.Context, monolithModule int, email, ott, firstName string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if firstName == "" {
		e["first_name"] = "First name is required"
	}
	if email == "" {
		e["email"] = "Email is required"
	}
	if ott == "" {
		e["ott"] = "One-time token is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for login OTT email", zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Send email
	//

	return uc.emailer.SendUserLoginOneTimeTokenEmail(ctx, monolithModule, email, ott, firstName)
}
