// cloud/mapleapps-backend/internal/maplefile/repo/collection/check.go
package collection

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM maplefile_collections_by_id_simplified WHERE id = ?`

	if err := impl.Session.Query(query, id).WithContext(ctx).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return count > 0, nil
}

func (impl *collectionRepositoryImpl) IsCollectionOwner(ctx context.Context, collectionID, userID gocql.UUID) (bool, error) {
	var accessType string

	query := `SELECT access_type FROM maplefile_collections_by_user_simplified
		WHERE user_id = ? AND collection_id = ?`

	err := impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&accessType)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}

	return accessType == "owner", nil
}

func (impl *collectionRepositoryImpl) CheckAccess(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error) {
	var accessType, permissionLevel string

	query := `SELECT access_type, permission_level FROM maplefile_collections_by_user_simplified
		WHERE user_id = ? AND collection_id = ?`

	err := impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&accessType, &permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil // No access
		}
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	// Owners have all permissions
	if accessType == "owner" {
		return true, nil
	}

	// Check if user's permission level meets requirement
	return impl.hasPermission(permissionLevel, requiredPermission), nil
}

func (impl *collectionRepositoryImpl) GetUserPermissionLevel(ctx context.Context, collectionID, userID gocql.UUID) (string, error) {
	var accessType, permissionLevel string

	query := `SELECT access_type, permission_level FROM maplefile_collections_by_user_simplified
		WHERE user_id = ? AND collection_id = ?`

	err := impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&accessType, &permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return "", nil // No access
		}
		return "", fmt.Errorf("failed to get permission level: %w", err)
	}

	if accessType == "owner" {
		return dom_collection.CollectionPermissionAdmin, nil
	}

	return permissionLevel, nil
}
