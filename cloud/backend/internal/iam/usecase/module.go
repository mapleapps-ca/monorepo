// cloud/backend/internal/iam/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/bannedipaddress"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/emailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			bannedipaddress.NewCreateBannedIPAddressUseCase,
			bannedipaddress.NewBannedIPAddressListAllValuesUseCase,
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
