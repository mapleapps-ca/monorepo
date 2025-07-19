// monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection/interface.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
)

// CollectionRepository defines the interface for collection persistence operations
type CollectionRepository interface {
	// Collection CRUD operations
	Create(ctx context.Context, collection *Collection) error
	Get(ctx context.Context, id gocql.UUID) (*Collection, error)
	Update(ctx context.Context, collection *Collection) error
	SoftDelete(ctx context.Context, id gocql.UUID) error // Now soft delete
	HardDelete(ctx context.Context, id gocql.UUID) error

	// State management operations
	Archive(ctx context.Context, id gocql.UUID) error
	Restore(ctx context.Context, id gocql.UUID) error

	// Hierarchical queries (now state-aware)
	FindByParent(ctx context.Context, parentID gocql.UUID) ([]*Collection, error)
	FindRootCollections(ctx context.Context, ownerID gocql.UUID) ([]*Collection, error)
	FindDescendants(ctx context.Context, collectionID gocql.UUID) ([]*Collection, error)
	// GetFullHierarchy(ctx context.Context, rootID gocql.UUID) (*Collection, error) // DEPRECATED AND WILL BE REMOVED

	// Move collection to a new parent
	MoveCollection(ctx context.Context, collectionID, newParentID gocql.UUID, updatedAncestors []gocql.UUID, updatedPathSegments []string) error

	// Collection ownership and access queries (now state-aware)
	CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error)
	GetAllByUserID(ctx context.Context, ownerID gocql.UUID) ([]*Collection, error)
	GetCollectionsSharedWithUser(ctx context.Context, userID gocql.UUID) ([]*Collection, error)
	IsCollectionOwner(ctx context.Context, collectionID, userID gocql.UUID) (bool, error)
	CheckAccess(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error)
	GetUserPermissionLevel(ctx context.Context, collectionID, userID gocql.UUID) (string, error)

	// Filtered collection queries (now state-aware)
	GetCollectionsWithFilter(ctx context.Context, options CollectionFilterOptions) (*CollectionFilterResult, error)

	// Collection membership operations
	AddMember(ctx context.Context, collectionID gocql.UUID, membership *CollectionMembership) error
	RemoveMember(ctx context.Context, collectionID, recipientID gocql.UUID) error
	UpdateMemberPermission(ctx context.Context, collectionID, recipientID gocql.UUID, newPermission string) error
	GetCollectionMembership(ctx context.Context, collectionID, recipientID gocql.UUID) (*CollectionMembership, error)

	// Hierarchical sharing
	AddMemberToHierarchy(ctx context.Context, rootID gocql.UUID, membership *CollectionMembership) error
	RemoveMemberFromHierarchy(ctx context.Context, rootID, recipientID gocql.UUID) error

	// GetCollectionSyncData retrieves collection sync data with pagination for the specified user
	GetCollectionSyncData(ctx context.Context, userID gocql.UUID, cursor *CollectionSyncCursor, limit int64) (*CollectionSyncResponse, error)
	GetCollectionSyncDataByAccessType(ctx context.Context, userID gocql.UUID, cursor *CollectionSyncCursor, limit int64, accessType string) (*CollectionSyncResponse, error)

	// Count operations for all collection types (folders + albums)
	CountOwnedCollections(ctx context.Context, userID gocql.UUID) (int, error)
	CountSharedCollections(ctx context.Context, userID gocql.UUID) (int, error)
	CountOwnedFolders(ctx context.Context, userID gocql.UUID) (int, error)
	CountSharedFolders(ctx context.Context, userID gocql.UUID) (int, error)
	CountTotalUniqueFolders(ctx context.Context, userID gocql.UUID) (int, error)
}
