// cloud/backend/internal/maplefile/repo/collection/get.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
)

func (impl *collectionRepositoryImpl) Get(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	collection, err := impl.getCollectionByID(ctx, id, true) // state-aware
	if err != nil {
		return nil, err
	}

	if collection == nil || collection.State != dom_collection.CollectionStateActive {
		return nil, nil // Collection not found or not active
	}

	return collection, nil
}

func (impl *collectionRepositoryImpl) GetWithAnyState(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	return impl.getCollectionByID(ctx, id, false) // state-agnostic
}

func (impl *collectionRepositoryImpl) getCollectionByID(ctx context.Context, id gocql.UUID, stateAware bool) (*dom_collection.Collection, error) {
	var (
		encryptedName, collectionType, encryptedKeyJSON      string
		membersJSON, ancestorIDsJSON                         string
		parentID, ownerID, createdByUserID, modifiedByUserID gocql.UUID
		createdAt, modifiedAt, tombstoneExpiry               time.Time
		version, tombstoneVersion                            uint64
		state                                                string
	)

	query := `SELECT id, owner_id, encrypted_name, collection_type, encrypted_collection_key,
		members, parent_id, ancestor_ids, created_at, created_by_user_id, modified_at,
		modified_by_user_id, version, state, tombstone_version, tombstone_expiry
		FROM maplefile_collections_by_id WHERE id = ?`

	// Use context properly in the query
	err := impl.Session.Query(query, id).WithContext(ctx).Scan(
		&id, &ownerID, &encryptedName, &collectionType, &encryptedKeyJSON,
		&membersJSON, &parentID, &ancestorIDsJSON, &createdAt, &createdByUserID,
		&modifiedAt, &modifiedByUserID, &version, &state, &tombstoneVersion, &tombstoneExpiry)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Apply state filtering if state-aware mode is enabled
	if stateAware && state != dom_collection.CollectionStateActive {
		return nil, nil // Collection exists but not in active state
	}

	// Deserialize complex fields
	members, err := impl.deserializeMembers(membersJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize members: %w", err)
	}

	ancestorIDs, err := impl.deserializeAncestorIDs(ancestorIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize ancestor IDs: %w", err)
	}

	encryptedKey, err := impl.deserializeEncryptedCollectionKey(encryptedKeyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize encrypted collection key: %w", err)
	}

	collection := &dom_collection.Collection{
		ID:                     id,
		OwnerID:                ownerID,
		EncryptedName:          encryptedName,
		CollectionType:         collectionType,
		EncryptedCollectionKey: encryptedKey,
		Members:                members,
		ParentID:               parentID,
		AncestorIDs:            ancestorIDs,
		CreatedAt:              createdAt,
		CreatedByUserID:        createdByUserID,
		ModifiedAt:             modifiedAt,
		ModifiedByUserID:       modifiedByUserID,
		Version:                version,
		State:                  state,
		TombstoneVersion:       tombstoneVersion,
		TombstoneExpiry:        tombstoneExpiry,
	}

	return collection, nil
}

func (impl *collectionRepositoryImpl) GetAllByUserID(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id
		WHERE owner_id = ? AND state = ?`

	iter := impl.Session.Query(query, ownerID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get collections by owner: %w", err)
	}

	// Fetch full collection details
	return impl.getCollectionsByIDs(ctx, collectionIDs)
}

func (impl *collectionRepositoryImpl) GetCollectionsSharedWithUser(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_recipient_id_and_collection_id
		WHERE recipient_id = ?`

	iter := impl.Session.Query(query, userID).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get shared collections: %w", err)
	}

	// Fetch full collection details and filter by state
	collections, err := impl.getCollectionsByIDs(ctx, collectionIDs)
	if err != nil {
		return nil, err
	}

	// Filter out non-active collections
	var activeCollections []*dom_collection.Collection
	for _, collection := range collections {
		if collection != nil && collection.State == dom_collection.CollectionStateActive {
			activeCollections = append(activeCollections, collection)
		}
	}

	return activeCollections, nil
}

func (impl *collectionRepositoryImpl) FindByParent(ctx context.Context, parentID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
		WHERE parent_id = ? AND state = ?`

	iter := impl.Session.Query(query, parentID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to find collections by parent: %w", err)
	}

	return impl.getCollectionsByIDs(ctx, collectionIDs)
}

func (impl *collectionRepositoryImpl) FindRootCollections(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	// Root collections have null parent_id
	query := `SELECT collection_id FROM maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id
		WHERE owner_id = ? AND state = ? AND parent_id = null ALLOW FILTERING`

	iter := impl.Session.Query(query, ownerID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to find root collections: %w", err)
	}

	return impl.getCollectionsByIDs(ctx, collectionIDs)
}

func (impl *collectionRepositoryImpl) FindDescendants(ctx context.Context, collectionID gocql.UUID) ([]*dom_collection.Collection, error) {
	var descendantIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
		WHERE ancestor_id = ? AND state = ?`

	iter := impl.Session.Query(query, collectionID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var descendantID gocql.UUID
	for iter.Scan(&descendantID) {
		descendantIDs = append(descendantIDs, descendantID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to find descendants: %w", err)
	}

	return impl.getCollectionsByIDs(ctx, descendantIDs)
}

// DEPRECATED AND WILL BE REMOVED
// func (impl *collectionRepositoryImpl) GetFullHierarchy(ctx context.Context, rootID gocql.UUID) (*dom_collection.Collection, error) {
// 	// Get root collection
// 	root, err := impl.Get(ctx, rootID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if root == nil {
// 		return nil, nil
// 	}

// 	// For now, return just the root. In a full implementation, you might want to
// 	// build the complete hierarchy tree structure
// 	return root, nil
// }

// Helper method to fetch multiple collections by IDs
func (impl *collectionRepositoryImpl) getCollectionsByIDs(ctx context.Context, ids []gocql.UUID) ([]*dom_collection.Collection, error) {
	if len(ids) == 0 {
		return []*dom_collection.Collection{}, nil
	}

	var collections []*dom_collection.Collection

	for _, id := range ids {
		collection, err := impl.getCollectionByID(ctx, id, false)
		if err != nil {
			impl.Logger.Warn("failed to get collection by ID",
				zap.String("collection_id", id.String()),
				zap.Error(err))
			continue
		}

		if collection != nil {
			collections = append(collections, collection)
		}
	}

	return collections, nil
}
