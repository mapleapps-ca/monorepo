// cloud/mapleapps-backend/internal/iam/usecase/module.go
package usecase

import (
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/emailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			emailer.NewSendFederatedUserPasswordResetEmailUseCase,
			emailer.NewSendFederatedUserVerificationEmailUseCase,
			emailer.NewSendLoginOTTEmailUseCase,
			federateduser.NewFederatedUserGetBySessionIDUseCase,
			federateduser.NewFederatedUserCountByFilterUseCase,
			federateduser.NewFederatedUserCreateUseCase,
			federateduser.NewFederatedUserDeleteFederatedUserByEmailUseCase,
			federateduser.NewFederatedUserDeleteByIDUseCase,
			federateduser.NewFederatedUserGetByEmailUseCase,
			federateduser.NewFederatedUserGetByIDUseCase,
			federateduser.NewFederatedUserGetByVerificationCodeUseCase,
			federateduser.NewFederatedUserListAllUseCase,
			federateduser.NewFederatedUserListByFilterUseCase,
			federateduser.NewFederatedUserUpdateUseCase,
		),
	)
}
