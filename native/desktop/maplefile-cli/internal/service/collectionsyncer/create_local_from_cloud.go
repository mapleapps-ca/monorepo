// internal/service/collectionsyncer/create_local_from_cloud.go
package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// CreateLocalCollectionFromCloudCollectionService defines the interface for creating a local collection from a cloud collection
type CreateLocalCollectionFromCloudCollectionService interface {
	Execute(ctx context.Context, cloudID gocql.UUID, password string) (*dom_collection.Collection, error)
}

// createLocalCollectionFromCloudCollectionService implements the CreateLocalCollectionFromCloudCollectionService interface
type createLocalCollectionFromCloudCollectionService struct {
	logger                      *zap.Logger
	cloudRepository             collectiondto.CollectionDTORepository
	localRepository             dom_collection.CollectionRepository
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
}

// NewCreateLocalCollectionFromCloudCollectionService creates a new use case for creating cloud collections
func NewCreateLocalCollectionFromCloudCollectionService(
	logger *zap.Logger,
	cloudRepository collectiondto.CollectionDTORepository,
	localRepository dom_collection.CollectionRepository,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
) CreateLocalCollectionFromCloudCollectionService {
	logger = logger.Named("CreateLocalCollectionFromCloudCollectionService")
	return &createLocalCollectionFromCloudCollectionService{
		logger:                      logger,
		cloudRepository:             cloudRepository,
		localRepository:             localRepository,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		collectionDecryptionService: collectionDecryptionService,
	}
}

// Execute creates a new cloud collection
func (uc *createLocalCollectionFromCloudCollectionService) Execute(ctx context.Context, cloudCollectionID gocql.UUID, password string) (*dom_collection.Collection, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if cloudCollectionID.IsZero() {
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
	// STEP 3: Submit our request to the cloud to get the collection details.
	//

	// Call the repository to get the collection
	cloudCollectionDTO, err := uc.cloudRepository.GetFromCloudByID(ctx, cloudCollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection from the cloud", err)
	}
	if cloudCollectionDTO == nil {
		err := errors.NewAppError("cloud collection not found", nil)
		uc.logger.Error("🚨 Failed to fetch collection from cloud",
			zap.Error(err))
		return nil, err
	}

	uc.logger.Debug("🔍 Cloud collection DTO debugging",
		zap.String("cloudCollectionID", cloudCollectionDTO.ID.Hex()),
		zap.String("cloudCollectionOwnerID", cloudCollectionDTO.OwnerID.Hex()),
		zap.String("cloudCollectionEncryptedName", cloudCollectionDTO.EncryptedName),
		zap.String("cloudCollectionType", cloudCollectionDTO.CollectionType),
		zap.String("cloudCollectionState", cloudCollectionDTO.State),
		zap.Int("cloudCollectionMembersCount", len(cloudCollectionDTO.Members)),
		zap.Any("cloudCollectionParentID", cloudCollectionDTO.ParentID))

	for i, memberDTO := range cloudCollectionDTO.Members {
		encryptedKeyLength := 0
		if memberDTO.EncryptedCollectionKey != nil {
			encryptedKeyLength = len(memberDTO.EncryptedCollectionKey.ToBoxSealBytes())
		}
		uc.logger.Debug("🔍 Cloud collection member DTO",
			zap.Int("memberIndex", i),
			zap.String("memberID", memberDTO.ID.Hex()),
			zap.String("recipientID", memberDTO.RecipientID.Hex()),
			zap.String("recipientEmail", memberDTO.RecipientEmail),
			zap.String("permissionLevel", memberDTO.PermissionLevel),
			zap.Bool("isInherited", memberDTO.IsInherited),
			zap.Int("encryptedKeyLength", encryptedKeyLength))
	}

	// ENHANCED DEBUGGING: Log current user info for comparison
	uc.logger.Debug("🔍 Current user info for comparison",
		zap.String("currentUserID", user.ID.Hex()),
		zap.String("currentUserEmail", user.Email),
		zap.String("currentUserName", user.Name))

	//
	// STEP 4: Perform any necessary validation before creating the local collection.
	//

	// Make sure the cloud collection hasn't been deleted.
	if cloudCollectionDTO.TombstoneVersion > 0 {
		uc.logger.Debug("⏭️ Skipping local collection creation from the cloud because it has been deleted",
			zap.String("id", cloudCollectionDTO.ID.Hex()))
		return nil, nil
	}

	//
	// STEP 5: Create a new collection domain object from the cloud data using a mapping function.
	//

	// Create a new collection domain object from the cloud data using a mapping function.
	newCollection := mapCollectionDTOToDomain(cloudCollectionDTO)

	//
	// STEP 6: Decrypt the collection with provided password
	//

	collectionKey, err := uc.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, newCollection, password)
	if err != nil {
		// ENHANCED HANDLING: Instead of failing completely, create the collection with encrypted name
		uc.logger.Warn("⚠️ Failed to decrypt collection key, creating collection with encrypted name only",
			zap.String("collectionID", cloudCollectionDTO.ID.Hex()),
			zap.Error(err))

		// Set a placeholder name to indicate encryption issue
		newCollection.Name = "[Encrypted - No Access]"

		// Still create the collection record for sync purposes, but mark it as problematic
		newCollection.SyncStatus = dom_collection.SyncStatusCloudOnly // Or create a new status

		// DEBUGGING: Log the issue for investigation
		uc.logger.Error("🚨 SYNC ISSUE: Collection accessible in sync but not decryptable",
			zap.String("collectionID", cloudCollectionDTO.ID.Hex()),
			zap.String("collectionOwnerID", cloudCollectionDTO.OwnerID.Hex()),
			zap.String("currentUserID", user.ID.Hex()),
			zap.Int("membersCount", len(cloudCollectionDTO.Members)),
			zap.String("suggestedAction", "Check backend API membership data for this collection"))

		// Execute the use case to create the local collection record with limited data
		if err := uc.localRepository.Create(ctx, newCollection); err != nil {
			uc.logger.Error("🚨 Failed to create new (local) collection from the cloud even with fallback",
				zap.String("id", cloudCollectionDTO.ID.Hex()),
				zap.Error(err))
			return nil, err
		}

		uc.logger.Info("⚠️ Created collection with limited access",
			zap.String("id", newCollection.ID.Hex()),
			zap.String("name", newCollection.Name))

		return newCollection, nil
	}
	defer crypto.ClearBytes(collectionKey)

	//
	// Step 7: Decrypt any encrypted collection data
	//
	collectionName, err := uc.collectionDecryptionService.ExecuteDecryptData(ctx, cloudCollectionDTO.EncryptedName, collectionKey)
	if err != nil {
		uc.logger.Error("failed to decrypt collection name", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt collection name", err)
	}
	if collectionName == "" {
		uc.logger.Error("failed to decrypt collection name - empty result", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt collection name", err)
	}
	newCollection.Name = collectionName

	uc.logger.Debug("🔍 Mapped local collection",
		zap.String("id", newCollection.ID.Hex()),
		zap.String("state", newCollection.State),
		zap.String("name", newCollection.Name), // This might be empty!
		zap.Any("parent_id", newCollection.ParentID),
		zap.Int("sync_status", int(newCollection.SyncStatus)))

	// Execute the use case to create the local collection record.
	if err := uc.localRepository.Create(ctx, newCollection); err != nil {
		uc.logger.Error("🚨 Failed to create new (local) collection from the cloud",
			zap.String("id", cloudCollectionDTO.ID.Hex()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 8: Return our local  collection response from the cloud.
	//

	return newCollection, nil
}

// mapMembersDTOToDomain maps a slice of uc_collectiondto.CollectionMembershipDTO to a slice of dom_collection.CollectionMembership.
// Updated to handle EncryptedCollectionKey as struct instead of []byte
func mapMembersDTOToDomain(membersDTO []*dom_collectiondto.CollectionMembershipDTO) []*dom_collection.CollectionMembership {
	if membersDTO == nil {
		return nil
	}
	members := make([]*dom_collection.CollectionMembership, len(membersDTO))
	for i, memberDTO := range membersDTO {
		if memberDTO != nil {
			members[i] = &dom_collection.CollectionMembership{
				ID:                     memberDTO.ID,
				CollectionID:           memberDTO.CollectionID,
				RecipientID:            memberDTO.RecipientID,
				RecipientEmail:         memberDTO.RecipientEmail,
				GrantedByID:            memberDTO.GrantedByID,
				EncryptedCollectionKey: memberDTO.EncryptedCollectionKey, // Now *keys.EncryptedCollectionKey
				PermissionLevel:        memberDTO.PermissionLevel,
				CreatedAt:              memberDTO.CreatedAt,
				IsInherited:            memberDTO.IsInherited,
				InheritedFromID:        memberDTO.InheritedFromID,
			}
		}
	}
	return members
}

// mapCollectionDTOToDomain maps a single uc_collectiondto.CollectionDTO to a single dom_collection.Collection.
// It handles nested Members and Children recursively by calling helper slice mappers.
func mapCollectionDTOToDomain(dto *dom_collectiondto.CollectionDTO) *dom_collection.Collection {
	if dto == nil {
		return nil
	}

	state := dto.State
	if state == "" {
		state = dom_collection.CollectionStateActive // Default to active
	}

	// Assuming dom_collection.Collection has fields compatible with uc_collectiondto.CollectionDTO
	return &dom_collection.Collection{
		ID:                     dto.ID,
		OwnerID:                dto.OwnerID,
		EncryptedName:          dto.EncryptedName,
		CollectionType:         dto.CollectionType,
		EncryptedCollectionKey: dto.EncryptedCollectionKey,         // Assuming keys.EncryptedCollectionKey type is compatible
		Members:                mapMembersDTOToDomain(dto.Members), // Call helper for members slice
		ParentID:               dto.ParentID,
		AncestorIDs:            dto.AncestorIDs,
		Children:               mapChildrenDTOToDomain(dto.Children), // Call helper for children slice
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,

		Name:       "[Encrypted]", // Placeholder until you implement decryption
		SyncStatus: dom_collection.SyncStatusSynced,

		State:            state,
		TombstoneVersion: dto.TombstoneVersion,
		TombstoneExpiry:  dto.TombstoneExpiry,
	}
}

// mapChildrenDTOToDomain maps a slice of uc_collectiondto.CollectionDTO to a slice of dom_collection.Collection.
// It recursively calls mapCollectionDTOToDomain for each child.
func mapChildrenDTOToDomain(childrenDTO []*dom_collectiondto.CollectionDTO) []*dom_collection.Collection {
	if childrenDTO == nil {
		return nil
	}
	children := make([]*dom_collection.Collection, len(childrenDTO))
	for i, childDTO := range childrenDTO {
		children[i] = mapCollectionDTOToDomain(childDTO) // Recursive call to single item mapper
	}
	return children
}
