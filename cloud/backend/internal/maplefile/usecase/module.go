// cloud/backend/internal/maplefile/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/bannedipaddress"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/emailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/user"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// Banned IP Address use cases
			bannedipaddress.NewCreateBannedIPAddressUseCase,
			bannedipaddress.NewBannedIPAddressListAllValuesUseCase,

			// Email use cases
			emailer.NewSendUserPasswordResetEmailUseCase,
			emailer.NewSendUserVerificationEmailUseCase,

			// User use cases
			user.NewUserCountByFilterUseCase,
			user.NewUserCreateUseCase,
			user.NewUserUpdateUseCase,
			user.NewUserListByFilterUseCase,
			user.NewUserListAllUseCase,
			user.NewUserGetByVerificationCodeUseCase,
			user.NewUserGetBySessionIDUseCase,
			user.NewUserGetByIDUseCase,
			user.NewUserGetByEmailUseCase,
			user.NewUserDeleteByIDUseCase,
			user.NewUserDeleteUserByEmailUseCase,

			// File use cases
			file.NewCreateFileUseCase,
			file.NewGetFileUseCase,
			file.NewListFilesByCollectionUseCase,
			file.NewUpdateFileUseCase,
			file.NewDeleteFileUseCase,
			file.NewStoreEncryptedDataUseCase,
			file.NewGetEncryptedDataUseCase,

			// Collection use cases
			collection.NewAddCollectionMemberUseCase,
			collection.NewAddMemberToHierarchyUseCase,
			collection.NewCheckCollectionAccessUseCase,
			collection.NewCreateCollectionUseCase,
			collection.NewDeleteCollectionUseCase,
			collection.NewFindCollectionsByParentUseCase,
			collection.NewFindDescendantsUseCase,
			collection.NewFindRootCollectionsUseCase,
			collection.NewGetCollectionUseCase,
			collection.NewGetCollectionHierarchyUseCase,
			collection.NewListCollectionsByUserUseCase,
			collection.NewListCollectionsSharedWithUserUseCase,
			collection.NewMoveCollectionUseCase,
			collection.NewRemoveCollectionMemberUseCase,
			collection.NewRemoveMemberFromHierarchyUseCase,
			collection.NewUpdateCollectionUseCase,
			collection.NewUpdateMemberPermissionUseCase,
		),
	)
}
