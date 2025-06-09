// cloud/backend/internal/maplefile/repo/collection/check.go
package collection

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM maplefile_collections_by_id WHERE id = ?`

	if err := impl.Session.Query(query, id).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return count > 0, nil
}

func (impl *collectionRepositoryImpl) IsCollectionOwner(ctx context.Context, collectionID, userID gocql.UUID) (bool, error) {
	collection, err := impl.GetWithAnyState(ctx, collectionID)
	if err != nil {
		return false, fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return false, nil
	}

	return collection.OwnerID == userID, nil
}

func (impl *collectionRepositoryImpl) CheckAccess(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error) {
	// Check if user is owner (owners have all permissions)
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	// Check membership
	var permissionLevel string

	query := `SELECT permission_level FROM maplefile_collections_by_recipient_id_and_collection_id
		WHERE recipient_id = ? AND collection_id = ?`

	err = impl.Session.Query(query, userID, collectionID).Scan(&permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil // No access
		}
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	// Check if user's permission level meets requirement
	return impl.hasPermission(permissionLevel, requiredPermission), nil
}

func (impl *collectionRepositoryImpl) GetUserPermissionLevel(ctx context.Context, collectionID, userID gocql.UUID) (string, error) {
	// Check if user is owner
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return "", err
	}

	if isOwner {
		return dom_collection.CollectionPermissionAdmin, nil
	}

	// Check membership
	var permissionLevel string

	query := `SELECT permission_level FROM maplefile_collections_by_recipient_id_and_collection_id
		WHERE recipient_id = ? AND collection_id = ?`

	err = impl.Session.Query(query, userID, collectionID).Scan(&permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return "", nil // No access
		}
		return "", fmt.Errorf("failed to get permission level: %w", err)
	}

	return permissionLevel, nil
}

// Helper method to check if a permission level meets requirements
func (impl *collectionRepositoryImpl) hasPermission(userPermission, requiredPermission string) bool {
	permissionLevels := map[string]int{
		dom_collection.CollectionPermissionReadOnly:  1,
		dom_collection.CollectionPermissionReadWrite: 2,
		dom_collection.CollectionPermissionAdmin:     3,
	}

	userLevel, userExists := permissionLevels[userPermission]
	requiredLevel, requiredExists := permissionLevels[requiredPermission]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
