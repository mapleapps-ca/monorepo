// internal/service/collectionsyncer/update_local_from_cloud.go
package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// UpdateLocalCollectionFromCloudCollectionService defines the interface for updating a local collection from a cloud collection
type UpdateLocalCollectionFromCloudCollectionService interface {
	Execute(ctx context.Context, cloudID gocql.UUID, password string) (*dom_collection.Collection, error)
}

// updateLocalCollectionFromCloudCollectionService implements the UpdateLocalCollectionFromCloudCollectionService interface
type updateLocalCollectionFromCloudCollectionService struct {
	logger                     *zap.Logger
	cloudRepository            collectiondto.CollectionDTORepository
	localRepository            dom_collection.CollectionRepository
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	decryptionService          collectioncrypto.CollectionDecryptionService
}

// NewUpdateLocalCollectionFromCloudCollectionService creates a new use case for updating local collection from the cloud
func NewUpdateLocalCollectionFromCloudCollectionService(
	logger *zap.Logger,
	cloudRepository collectiondto.CollectionDTORepository,
	localRepository dom_collection.CollectionRepository,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	decryptionService collectioncrypto.CollectionDecryptionService,
) UpdateLocalCollectionFromCloudCollectionService {
	logger = logger.Named("UpdateLocalCollectionFromCloudCollectionService")
	return &updateLocalCollectionFromCloudCollectionService{
		logger:                     logger,
		cloudRepository:            cloudRepository,
		localRepository:            localRepository,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		decryptionService:          decryptionService,
	}
}

// Execute updates a local collection from the cloud
func (uc *updateLocalCollectionFromCloudCollectionService) Execute(ctx context.Context, cloudCollectionID gocql.UUID, password string) (*dom_collection.Collection, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if cloudCollectionID.String() == "" {
		e["cloudCollectionID"] = "Cloud ID is required"
	}
	if password == "" {
		e["password"] = "Password is required"
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// Step 2: Get user and collection for E2EE key chain
	//

	user, err := uc.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("user not found", nil)
	}

	//
	// STEP 3: Submit our request to the cloud to get the collection details and get related local collections.
	//

	// Call the repository to get the collection
	cloudCollectionDTO, err := uc.cloudRepository.GetFromCloudByID(ctx, cloudCollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection from the cloud", err)
	}
	if cloudCollectionDTO == nil {
		err := errors.NewAppError("cloud collection not found", nil)
		uc.logger.Error("❌ Failed to fetch collection from cloud",
			zap.Error(err))
		return nil, err
	}

	// Call the repository to get the related local collections
	localCollection, err := uc.localRepository.GetByID(ctx, cloudCollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local collections", err)
	}

	//
	// STEP 4: Confirm that the local collection exists.
	//
	if localCollection == nil {
		err := errors.NewAppError("no local collection found", nil)
		uc.logger.Error("❌ Failed to fetch local collections",
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Confirm we are eligible for updating the local collection.
	//

	// CASE 1: Local collection is already same or newest version compared with the cloud collection.
	if localCollection.Version >= cloudCollectionDTO.Version {
		uc.logger.Debug("✅ Local collection is already same or newest version compared with the cloud collection",
			zap.String("collection_id", cloudCollectionID.String()))
		return nil, nil
	}
	// CASE 2: We must handle local deletion of the collection.
	if cloudCollectionDTO.TombstoneVersion > localCollection.Version {
		if err := uc.localRepository.Delete(ctx, localCollection.ID); err != nil {
			uc.logger.Error("❌ Failed to delete local collection",
				zap.String("collection_id", cloudCollectionID.String()),
				zap.Uint64("local_version", localCollection.Version),
				zap.Uint64("cloud_version", cloudCollectionDTO.Version),
				zap.Error(err))
			return nil, err
		}
		uc.logger.Debug("🗑️ Local collection is marked as deleted",
			zap.String("collection_id", cloudCollectionID.String()),
			zap.Uint64("local_version", localCollection.Version),
			zap.Uint64("cloud_version", cloudCollectionDTO.Version))
		return nil, nil
	}

	//
	// STEP 6: Decrypt the collection with provided password
	//

	collectionKey, err := uc.decryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, localCollection, password)
	if err != nil {
		uc.logger.Warn("⚠️ Failed to decrypt collection key, using placeholder",
			zap.String("collectionID", cloudCollectionDTO.ID.String()),
			zap.Error(err))
		return nil, err
	}
	defer crypto.ClearBytes(collectionKey)

	//
	// Step 7: Decrypt any encrypted collection data
	//
	collectionName, err := uc.decryptionService.ExecuteDecryptData(ctx, cloudCollectionDTO.EncryptedName, collectionKey)
	if err != nil {
		uc.logger.Error("failed to decrypt file metadata", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt file metadata", err)
	}
	if collectionName == "" {
		uc.logger.Error("failed to decrypt collection name", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt collection name", err)
	}

	//
	// STEP 6: Update the local existing collection from the cloud.
	//

	// Update a new collection domain object from the cloud data using a mapping function.
	cloudCollection := mapCollectionDTOToDomain(cloudCollectionDTO)

	// IMPORTANT: Assign our decrypted values to.
	cloudCollection.Name = collectionName

	uc.logger.Debug("🔍 Full cloud collection DTO",
		zap.String("id", cloudCollectionDTO.ID.String()),
		zap.String("state", cloudCollectionDTO.State),
		zap.String("encrypted_name", cloudCollectionDTO.EncryptedName),
		zap.Any("parent_id", cloudCollectionDTO.ParentID))

	// Execute the use case to update the local collection record.
	if err := uc.localRepository.Save(ctx, cloudCollection); err != nil {
		uc.logger.Error("❌ Failed to update new (local) collection from the cloud",
			zap.String("id", cloudCollectionDTO.ID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 6: Return our local  collection response from the cloud.
	//

	uc.logger.Debug("✅ Local collection is updated",
		zap.String("id", cloudCollection.ID.String()),
		zap.String("name", cloudCollection.Name),
	)
	return cloudCollection, nil
}
