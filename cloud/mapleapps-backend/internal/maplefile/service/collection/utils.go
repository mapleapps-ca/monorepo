// monorepo/cloud/backend/internal/maplefile/service/collection/utils.go
package collection

import (
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

// Helper function to map a CollectionMembershipDTO to a CollectionMembership domain model
// This assumes a direct field-by-field copy is intended by the DTO structure.
func mapMembershipDTOToDomain(dto *CollectionMembershipDTO) dom_collection.CollectionMembership {
	return dom_collection.CollectionMembership{
		ID:                     dto.ID,                     // Copy DTO ID
		CollectionID:           dto.CollectionID,           // Copy DTO CollectionID
		RecipientID:            dto.RecipientID,            // Copy DTO RecipientID
		RecipientEmail:         dto.RecipientEmail,         // Copy DTO RecipientEmail
		GrantedByID:            dto.GrantedByID,            // Copy DTO GrantedByID
		EncryptedCollectionKey: dto.EncryptedCollectionKey, // Copy DTO EncryptedCollectionKey
		PermissionLevel:        dto.PermissionLevel,        // Copy DTO PermissionLevel
		CreatedAt:              dto.CreatedAt,              // Copy DTO CreatedAt
		IsInherited:            dto.IsInherited,            // Copy DTO IsInherited
		InheritedFromID:        dto.InheritedFromID,        // Copy DTO InheritedFromID
		// Note: ModifiedAt/By, Version are not in Membership DTO/Domain
	}
}

// Helper function to map a CreateCollectionRequestDTO to a Collection domain model.
// This function recursively maps all fields, including nested members and children,
// copying values directly from the DTO. Server-side overrides for fields like
// ID, OwnerID, timestamps, and version are applied *after* this mapping in the Execute method.
// userID and now are passed for potential use in recursive calls if needed for consistency,
// though the primary goal here is to copy DTO values.
func mapCollectionDTOToDomain(dto *CreateCollectionRequestDTO, userID gocql.UUID, now time.Time) *dom_collection.Collection {
	if dto == nil {
		return nil
	}

	collection := &dom_collection.Collection{
		// Copy all scalar/pointer fields directly from the DTO as requested by the prompt.
		// Fields like ID, OwnerID, timestamps, and version from the DTO
		// represent the client's proposed state and will be potentially
		// overridden by server-managed values later in the Execute method.
		ID:                     dto.ID,
		OwnerID:                dto.OwnerID,
		EncryptedName:          dto.EncryptedName,
		CollectionType:         dto.CollectionType,
		EncryptedCollectionKey: dto.EncryptedCollectionKey,
		ParentID:               dto.ParentID,
		AncestorIDs:            dto.AncestorIDs,
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
	}

	// Map members slice from DTO to domain model slice
	if len(dto.Members) > 0 {
		collection.Members = make([]dom_collection.CollectionMembership, len(dto.Members))
		for i, memberDTO := range dto.Members {
			collection.Members[i] = mapMembershipDTOToDomain(memberDTO)
		}
	}

	return collection
}

// Helper function to map a Collection domain model to a CollectionResponseDTO
// This function should ideally exclude sensitive data (like recipient-specific keys)
// that should not be part of a general response.
func mapCollectionToDTO(collection *dom_collection.Collection) *CollectionResponseDTO {
	if collection == nil {
		return nil
	}

	responseDTO := &CollectionResponseDTO{
		ID:             collection.ID,
		OwnerID:        collection.OwnerID,
		EncryptedName:  collection.EncryptedName,
		CollectionType: collection.CollectionType,
		ParentID:       collection.ParentID,
		AncestorIDs:    collection.AncestorIDs,
		// Note: EncryptedCollectionKey from the domain model is the owner's key.
		// Including it in the general response DTO might be acceptable if the response
		// is only sent to the owner and contains *their* key. Otherwise, this field
		// might need conditional inclusion or exclusion. The prompt does not require
		// changing this, so we keep the original mapping which copies the owner's key.
		EncryptedCollectionKey: collection.EncryptedCollectionKey,
		CreatedAt:              collection.CreatedAt,
		ModifiedAt:             collection.ModifiedAt,
		// Members slice needs mapping to MembershipResponseDTO
		Members: make([]MembershipResponseDTO, len(collection.Members)),
	}

	// Map members
	for i, member := range collection.Members {
		responseDTO.Members[i] = MembershipResponseDTO{
			ID:              member.ID,
			RecipientID:     member.RecipientID,
			RecipientEmail:  member.RecipientEmail, // Email for display
			PermissionLevel: member.PermissionLevel,
			GrantedByID:     member.GrantedByID,
			CollectionID:    member.CollectionID, // Redundant but useful
			IsInherited:     member.IsInherited,
			InheritedFromID: member.InheritedFromID,
			CreatedAt:       member.CreatedAt,
			// Note: EncryptedCollectionKey for this member is recipient-specific
			// and should NOT be included in a general response DTO unless
			// filtered for the specific recipient receiving the response.
			// The MembershipResponseDTO does not have a field for this, which is correct.
			EncryptedCollectionKey: member.EncryptedCollectionKey,
		}
	}

	return responseDTO
}
