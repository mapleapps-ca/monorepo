package usecase

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/emailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// Email use cases
			emailer.NewSendUserPasswordResetEmailUseCase,
			emailer.NewSendUserVerificationEmailUseCase,

			// User use cases
			user.NewUserCreateUseCase,
			user.NewUserUpdateUseCase,
			user.NewUserGetByVerificationCodeUseCase,
			user.NewUserGetBySessionIDUseCase,
			user.NewUserGetByIDUseCase,
			user.NewUserGetByEmailUseCase,
			user.NewUserDeleteByIDUseCase,
			user.NewUserDeleteUserByEmailUseCase,

			// // Collection use cases
			collection.NewAddCollectionMemberUseCase,
			collection.NewAddMemberToHierarchyUseCase,
			collection.NewCheckCollectionAccessUseCase,
			collection.NewCreateCollectionUseCase,
			collection.NewSoftDeleteCollectionUseCase,
			collection.NewFindCollectionsByParentUseCase,
			collection.NewFindDescendantsUseCase,
			collection.NewFindRootCollectionsUseCase,
			collection.NewGetCollectionUseCase,
			collection.NewGetFilteredCollectionsUseCase,
			collection.NewListCollectionsByUserUseCase,
			collection.NewListCollectionsSharedWithUserUseCase,
			collection.NewMoveCollectionUseCase,
			collection.NewRemoveCollectionMemberUseCase,
			collection.NewRemoveMemberFromHierarchyUseCase,
			collection.NewUpdateCollectionUseCase,
			collection.NewUpdateMemberPermissionUseCase,
			collection.NewGetCollectionSyncDataUseCase,
			collection.NewCountUserCollectionsUseCase,

			// File Metadata use cases
			filemetadata.NewCreateFileMetadataUseCase,
			filemetadata.NewCreateManyFileMetadataUseCase,
			filemetadata.NewGetFileMetadataUseCase,
			filemetadata.NewGetFileMetadataByIDsUseCase,
			filemetadata.NewGetFileMetadataByCollectionUseCase,
			filemetadata.NewUpdateFileMetadataUseCase,
			filemetadata.NewSoftDeleteFileMetadataUseCase,
			filemetadata.NewDeleteManyFileMetadataUseCase,
			filemetadata.NewCheckFileExistsUseCase,
			filemetadata.NewCheckFileAccessUseCase,
			filemetadata.NewGetFileMetadataByCreatedByUserIDUseCase,
			filemetadata.NewGetFileMetadataByOwnerIDUseCase,
			filemetadata.NewListFileMetadataSyncDataUseCase,
			filemetadata.NewCountUserFilesUseCase,
			filemetadata.NewGetStorageSizeByOwnerUseCase,
			filemetadata.NewGetStorageSizeByUserUseCase,
			filemetadata.NewGetStorageSizeByCollectionUseCase,

			// File Object Storage use cases
			fileobjectstorage.NewStoreEncryptedDataUseCase,
			fileobjectstorage.NewGetEncryptedDataUseCase,
			fileobjectstorage.NewDeleteEncryptedDataUseCase,
			fileobjectstorage.NewDeleteMultipleEncryptedDataUseCase,
			fileobjectstorage.NewStoreMultipleEncryptedDataUseCase,
			fileobjectstorage.NewGeneratePresignedUploadURLUseCase,
			fileobjectstorage.NewGeneratePresignedDownloadURLUseCase,
			fileobjectstorage.NewVerifyObjectExistsUseCase,
			fileobjectstorage.NewGetObjectSizeUseCase,
		),
	)
}
