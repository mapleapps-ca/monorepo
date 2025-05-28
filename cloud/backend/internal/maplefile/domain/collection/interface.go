// monorepo/cloud/backend/internal/maplefile/domain/collection/repository.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionRepository defines the interface for collection persistence operations
type CollectionRepository interface {
	// Collection CRUD operations
	Create(ctx context.Context, collection *Collection) error
	Get(ctx context.Context, id primitive.ObjectID) (*Collection, error)
	GetWithAnyState(ctx context.Context, id primitive.ObjectID) (*Collection, error)
	Update(ctx context.Context, collection *Collection) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error // Now soft delete
	HardDelete(ctx context.Context, id primitive.ObjectID) error

	// State management operations
	Archive(ctx context.Context, id primitive.ObjectID) error
	Restore(ctx context.Context, id primitive.ObjectID) error

	// Hierarchical queries (now state-aware)
	FindByParent(ctx context.Context, parentID primitive.ObjectID) ([]*Collection, error)
	FindRootCollections(ctx context.Context, ownerID primitive.ObjectID) ([]*Collection, error)
	FindDescendants(ctx context.Context, collectionID primitive.ObjectID) ([]*Collection, error)
	GetFullHierarchy(ctx context.Context, rootID primitive.ObjectID) (*Collection, error)

	// Move collection to a new parent
	MoveCollection(ctx context.Context, collectionID, newParentID primitive.ObjectID, updatedAncestors []primitive.ObjectID, updatedPathSegments []string) error

	// Collection ownership and access queries (now state-aware)
	CheckIfExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error)
	GetAllByUserID(ctx context.Context, ownerID primitive.ObjectID) ([]*Collection, error)
	GetCollectionsSharedWithUser(ctx context.Context, userID primitive.ObjectID) ([]*Collection, error)
	IsCollectionOwner(ctx context.Context, collectionID, userID primitive.ObjectID) (bool, error)
	CheckAccess(ctx context.Context, collectionID, userID primitive.ObjectID, requiredPermission string) (bool, error)
	GetUserPermissionLevel(ctx context.Context, collectionID, userID primitive.ObjectID) (string, error)

	// Filtered collection queries (now state-aware)
	GetCollectionsWithFilter(ctx context.Context, options CollectionFilterOptions) (*CollectionFilterResult, error)

	// Collection membership operations
	AddMember(ctx context.Context, collectionID primitive.ObjectID, membership *CollectionMembership) error
	RemoveMember(ctx context.Context, collectionID, recipientID primitive.ObjectID) error
	UpdateMemberPermission(ctx context.Context, collectionID, recipientID primitive.ObjectID, newPermission string) error
	GetCollectionMembership(ctx context.Context, collectionID, recipientID primitive.ObjectID) (*CollectionMembership, error)

	// Hierarchical sharing
	AddMemberToHierarchy(ctx context.Context, rootID primitive.ObjectID, membership *CollectionMembership) error
	RemoveMemberFromHierarchy(ctx context.Context, rootID, recipientID primitive.ObjectID) error

	// GetCollectionSyncData retrieves collection sync data with pagination for the specified user
	GetCollectionSyncData(ctx context.Context, userID primitive.ObjectID, cursor *CollectionSyncCursor, limit int64) (*CollectionSyncResponse, error)
}
