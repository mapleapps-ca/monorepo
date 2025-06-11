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

	query := `SELECT COUNT(*) FROM maplefile_collections_by_id WHERE id = ?`

	if err := impl.Session.Query(query, id).WithContext(ctx).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return count > 0, nil
}

// IsCollectionOwner demonstrates the memory-filtering approach for better performance
// Instead of forcing Cassandra to scan with ALLOW FILTERING, we query efficiently and filter in memory
func (impl *collectionRepositoryImpl) IsCollectionOwner(ctx context.Context, collectionID, userID gocql.UUID) (bool, error) {
	// Strategy: Use the compound partition key table to efficiently check ownership
	// This query is fast because both user_id and access_type are part of the partition key
	var collectionExists gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'owner' AND collection_id = ? LIMIT 1 ALLOW FILTERING`

	err := impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&collectionExists)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}

	// If we got a result, the user is an owner of this collection
	return true, nil
}

// Alternative implementation using the memory-filtering approach
// This demonstrates a different strategy when you can't avoid some filtering
func (impl *collectionRepositoryImpl) IsCollectionOwnerAlternative(ctx context.Context, collectionID, userID gocql.UUID) (bool, error) {
	// Memory-filtering approach: Get all collections for this user, filter for the specific collection
	// This is efficient when users don't have thousands of collections

	query := `SELECT collection_id, access_type FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ?`

	iter := impl.Session.Query(query, userID).WithContext(ctx).Iter()

	var currentCollectionID gocql.UUID
	var accessType string

	for iter.Scan(&currentCollectionID, &accessType) {
		// Check if this is the collection we're looking for and if the user is the owner
		if currentCollectionID == collectionID && accessType == "owner" {
			iter.Close()
			return true, nil
		}
	}

	if err := iter.Close(); err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}

	return false, nil
}

// CheckAccess uses the efficient compound partition key approach
func (impl *collectionRepositoryImpl) CheckAccess(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error) {
	// First check if user is owner (owners have all permissions)
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}

	if isOwner {
		return true, nil // Owners have all permissions
	}

	// Check if user is a member with sufficient permissions
	var permissionLevel string

	query := `SELECT permission_level FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'member' AND collection_id = ? LIMIT 1 ALLOW FILTERING`

	err = impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil // No access
		}
		return false, fmt.Errorf("failed to check member access: %w", err)
	}

	// Check if user's permission level meets requirement
	return impl.hasPermission(permissionLevel, requiredPermission), nil
}

// GetUserPermissionLevel efficiently determines a user's permission level for a collection
func (impl *collectionRepositoryImpl) GetUserPermissionLevel(ctx context.Context, collectionID, userID gocql.UUID) (string, error) {
	// Check ownership first using the efficient compound key table
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to check ownership: %w", err)
	}

	if isOwner {
		return dom_collection.CollectionPermissionAdmin, nil
	}

	// Check member permissions
	var permissionLevel string

	query := `SELECT permission_level FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'member' AND collection_id = ? LIMIT 1 ALLOW FILTERING`

	err = impl.Session.Query(query, userID, collectionID).WithContext(ctx).Scan(&permissionLevel)
	if err != nil {
		if err == gocql.ErrNotFound {
			return "", nil // No access
		}
		return "", fmt.Errorf("failed to get permission level: %w", err)
	}

	return permissionLevel, nil
}

// Demonstration of a completely ALLOW FILTERING-free approach using direct collection lookup
// This approach queries the main collection table and checks ownership directly
func (impl *collectionRepositoryImpl) CheckAccessByCollectionLookup(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error) {
	// Strategy: Get the collection directly and check ownership/membership from the collection object
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return false, fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return false, nil // Collection doesn't exist
	}

	// Check if user is the owner
	if collection.OwnerID == userID {
		return true, nil // Owners have all permissions
	}

	// Check if user is a member with sufficient permissions
	for _, member := range collection.Members {
		if member.RecipientID == userID {
			return impl.hasPermission(member.PermissionLevel, requiredPermission), nil
		}
	}

	return false, nil // User has no access
}
