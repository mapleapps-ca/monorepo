// cloud/mapleapps-backend/internal/iam/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/emailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/recovery"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			emailer.NewSendFederatedUserPasswordResetEmailUseCase,
			emailer.NewSendFederatedUserVerificationEmailUseCase,
			emailer.NewSendLoginOTTEmailUseCase,
			federateduser.NewFederatedUserGetBySessionIDUseCase,
			federateduser.NewFederatedUserCreateUseCase,
			federateduser.NewFederatedUserDeleteFederatedUserByEmailUseCase,
			federateduser.NewFederatedUserDeleteByIDUseCase,
			federateduser.NewFederatedUserGetByEmailUseCase,
			federateduser.NewFederatedUserGetByIDUseCase,
			federateduser.NewFederatedUserGetByVerificationCodeUseCase,
			federateduser.NewFederatedUserUpdateUseCase,
			recovery.NewInitiateRecoveryUseCase,
			recovery.NewVerifyRecoveryUseCase,
			recovery.NewCompleteRecoveryUseCase,
		),
	)
}
