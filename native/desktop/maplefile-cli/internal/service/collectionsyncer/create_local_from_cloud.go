// internal/service/collectionsyncer/create_local_from_cloud.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// CreateLocalCollectionFromCloudCollectionService defines the interface for creating a local collection from a cloud collection
type CreateLocalCollectionFromCloudCollectionService interface {
	Execute(ctx context.Context, cloudID primitive.ObjectID) (*dom_collection.Collection, error)
}

// createLocalCollectionFromCloudCollectionService implements the CreateLocalCollectionFromCloudCollectionService interface
type createLocalCollectionFromCloudCollectionService struct {
	logger          *zap.Logger
	cloudRepository collectiondto.CollectionDTORepository
	localRepository dom_collection.CollectionRepository
}

// NewCreateLocalCollectionFromCloudCollectionService creates a new use case for creating cloud collections
func NewCreateLocalCollectionFromCloudCollectionService(
	logger *zap.Logger,
	cloudRepository collectiondto.CollectionDTORepository,
	localRepository dom_collection.CollectionRepository,
) CreateLocalCollectionFromCloudCollectionService {
	return &createLocalCollectionFromCloudCollectionService{
		logger:          logger,
		cloudRepository: cloudRepository,
		localRepository: localRepository,
	}
}

// Execute creates a new cloud collection
func (uc *createLocalCollectionFromCloudCollectionService) Execute(ctx context.Context, cloudCollectionID primitive.ObjectID) (*dom_collection.Collection, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if cloudCollectionID.IsZero() {
		e["cloudCollectionID"] = "Cloud ID is required"
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Submit our request to the cloud to get the collection details.
	//

	// Call the repository to get the collection
	cloudCollectionDTO, err := uc.cloudRepository.GetFromCloudByID(ctx, cloudCollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection from the cloud", err)
	}
	if cloudCollectionDTO == nil {
		err := errors.NewAppError("cloud collection not found", nil)
		uc.logger.Error("Failed to fetch collection from cloud",
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Perform any necessary validation before creating the local collection.
	//

	// CASE 1: Make sure the cloud collection hasn't been deleted.
	if cloudCollectionDTO.TombstoneVersion > 0 {
		uc.logger.Debug("Skipping local collection creation from the cloud because it has been deleted",
			zap.String("id", cloudCollectionDTO.ID.Hex()))
		return nil, nil
	}

	//
	// STEP 4: Create a new collection domain object from the cloud data using a mapping function.
	//

	// Create a new collection domain object from the cloud data using a mapping function.
	newCollection := mapCollectionDTOToDomain(cloudCollectionDTO)

	// Execute the use case to create the local collection record.
	if err := uc.localRepository.Create(ctx, newCollection); err != nil {
		uc.logger.Error("Failed to create new (local) collection from the cloud",
			zap.String("id", cloudCollectionDTO.ID.Hex()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Return our local  collection response from the cloud.
	//

	return newCollection, nil
}

// mapMembersDTOToDomain maps a slice of uc_collectiondto.CollectionMembershipDTO to a slice of dom_collection.CollectionMembership.
// Assuming dom_collection.CollectionMembership struct matches uc_collectiondto.CollectionMembershipDTO.
func mapMembersDTOToDomain(membersDTO []*dom_collectiondto.CollectionMembershipDTO) []*dom_collection.CollectionMembership {
	if membersDTO == nil {
		return nil
	}
	members := make([]*dom_collection.CollectionMembership, len(membersDTO))
	for i, memberDTO := range membersDTO {
		if memberDTO != nil {
			// Assuming dom_collection.CollectionMembership has fields matching uc_collectiondto.CollectionMembershipDTO
			members[i] = &dom_collection.CollectionMembership{
				ID:                     memberDTO.ID,
				CollectionID:           memberDTO.CollectionID,
				RecipientID:            memberDTO.RecipientID,
				RecipientEmail:         memberDTO.RecipientEmail,
				GrantedByID:            memberDTO.GrantedByID,
				EncryptedCollectionKey: memberDTO.EncryptedCollectionKey,
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
		State:                  dto.State,
		// Note: TombstoneVersion and TombstoneExpiry from sync item are not part of the main CollectionDTO
		// and are not mapped here into the domain object. They are specific to the sync process.
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
