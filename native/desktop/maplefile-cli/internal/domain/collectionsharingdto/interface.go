// internal/domain/collection/interface.go
package collectionsharingdto

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

// CollectionSharingDTORepository defines the interface for collection sharing operations
type CollectionSharingDTORepository interface {
	// ShareCollectionInCloud shares a collection with another user
	ShareCollectionInCloud(ctx context.Context, request *ShareCollectionRequestDTO) (*ShareCollectionResponseDTO, error)

	// RemoveMemberInCloud removes a user's access to a collection
	RemoveMemberInCloud(ctx context.Context, request *RemoveMemberRequestDTO) (*RemoveMemberResponseDTO, error)

	// GetCollectionWithMembersFromCloud retrieves collection information including member list
	GetCollectionWithMembersFromCloud(ctx context.Context, collectionID gocql.UUID) (*collectiondto.CollectionDTO, error)

	// ListSharedCollectionsFromCloud gets all collections that have been shared with the authenticated user
	ListSharedCollectionsFromCloud(ctx context.Context) ([]*collectiondto.CollectionDTO, error)
}

// ValidatePermissionLevel validates that the permission level is valid
func ValidatePermissionLevel(permissionLevel string) error {
	switch permissionLevel {
	case CollectionDTOPermissionReadOnly, CollectionDTOPermissionReadWrite, CollectionDTOPermissionAdmin:
		return nil
	default:
		return fmt.Errorf("invalid permission level: %s (must be one of: %s, %s, %s)",
			permissionLevel, CollectionDTOPermissionReadOnly, CollectionDTOPermissionReadWrite, CollectionDTOPermissionAdmin)
	}
}

// CanUserShareCollection checks if a user has permission to share a collection
func CanUserShareCollection(membership *collectiondto.CollectionMembershipDTO) bool {
	if membership == nil {
		return false
	}
	// Only admin members can share collections
	return membership.PermissionLevel == CollectionDTOPermissionAdmin
}

// CanUserRemoveMembers checks if a user has permission to remove members from a collection
func CanUserRemoveMembers(membership *collectiondto.CollectionMembershipDTO) bool {
	if membership == nil {
		return false
	}
	// Only admin members can remove other members
	return membership.PermissionLevel == CollectionDTOPermissionAdmin
}
