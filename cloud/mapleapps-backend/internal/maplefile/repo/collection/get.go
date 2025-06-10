// cloud/mapleapps-backend/internal/maplefile/repo/collection/get.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
)

// Core helper methods for loading collections with members
func (impl *collectionRepositoryImpl) loadCollectionWithMembers(ctx context.Context, collectionID gocql.UUID, stateAware bool) (*dom_collection.Collection, error) {
	// 1. Load base collection
	collection, err := impl.getBaseCollection(ctx, collectionID, stateAware)
	if err != nil || collection == nil {
		return collection, err
	}

	// 2. Load and populate members
	members, err := impl.getCollectionMembers(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	collection.Members = members

	return collection, nil
}

func (impl *collectionRepositoryImpl) getBaseCollection(ctx context.Context, id gocql.UUID, stateAware bool) (*dom_collection.Collection, error) {
	var (
		encryptedName, collectionType, encryptedKeyJSON      string
		ancestorIDsJSON                                      string
		parentID, ownerID, createdByUserID, modifiedByUserID gocql.UUID
		createdAt, modifiedAt, tombstoneExpiry               time.Time
		version, tombstoneVersion                            uint64
		state                                                string
	)

	query := `SELECT id, owner_id, encrypted_name, collection_type, encrypted_collection_key,
		parent_id, ancestor_ids, created_at, created_by_user_id, modified_at,
		modified_by_user_id, version, state, tombstone_version, tombstone_expiry
		FROM maplefile_collections_by_id WHERE id = ?`

	err := impl.Session.Query(query, id).WithContext(ctx).Scan(
		&id, &ownerID, &encryptedName, &collectionType, &encryptedKeyJSON,
		&parentID, &ancestorIDsJSON, &createdAt, &createdByUserID,
		&modifiedAt, &modifiedByUserID, &version, &state, &tombstoneVersion, &tombstoneExpiry)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Apply state filtering if state-aware mode is enabled
	if stateAware && state != dom_collection.CollectionStateActive {
		return nil, nil
	}

	// Deserialize complex fields
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
		Members:                []dom_collection.CollectionMembership{}, // Will be populated separately
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

func (impl *collectionRepositoryImpl) getCollectionMembers(ctx context.Context, collectionID gocql.UUID) ([]dom_collection.CollectionMembership, error) {
	var members []dom_collection.CollectionMembership

	query := `SELECT recipient_id, member_id, recipient_email, granted_by_id,
		encrypted_collection_key, permission_level, created_at,
		is_inherited, inherited_from_id
		FROM maplefile_collection_members_by_collection_id_and_recipient_id WHERE collection_id = ?`

	iter := impl.Session.Query(query, collectionID).WithContext(ctx).Iter()

	var (
		recipientID, memberID, grantedByID, inheritedFromID gocql.UUID
		recipientEmail, permissionLevel                     string
		encryptedCollectionKey                              []byte
		createdAt                                           time.Time
		isInherited                                         bool
	)

	for iter.Scan(&recipientID, &memberID, &recipientEmail, &grantedByID,
		&encryptedCollectionKey, &permissionLevel, &createdAt,
		&isInherited, &inheritedFromID) {

		member := dom_collection.CollectionMembership{
			ID:                     memberID,
			CollectionID:           collectionID,
			RecipientID:            recipientID,
			RecipientEmail:         recipientEmail,
			GrantedByID:            grantedByID,
			EncryptedCollectionKey: encryptedCollectionKey,
			PermissionLevel:        permissionLevel,
			CreatedAt:              createdAt,
			IsInherited:            isInherited,
			InheritedFromID:        inheritedFromID,
		}
		members = append(members, member)
	}

	return members, iter.Close()
}

func (impl *collectionRepositoryImpl) loadMultipleCollectionsWithMembers(ctx context.Context, collectionIDs []gocql.UUID) ([]*dom_collection.Collection, error) {
	if len(collectionIDs) == 0 {
		return []*dom_collection.Collection{}, nil
	}

	var collections []*dom_collection.Collection
	for _, id := range collectionIDs {
		collection, err := impl.loadCollectionWithMembers(ctx, id, true)
		if err != nil {
			impl.Logger.Warn("failed to load collection",
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

func (impl *collectionRepositoryImpl) Get(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	return impl.loadCollectionWithMembers(ctx, id, true) // state-aware
}

func (impl *collectionRepositoryImpl) GetWithAnyState(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	return impl.loadCollectionWithMembers(ctx, id, false) // state-agnostic
}

func (impl *collectionRepositoryImpl) GetAllByUserID(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'owner' AND state = ?`

	iter := impl.Session.Query(query, ownerID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get collections by owner: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}

func (impl *collectionRepositoryImpl) GetCollectionsSharedWithUser(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	query := `SELECT collection_id FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'member' AND state = ?`

	iter := impl.Session.Query(query, userID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get shared collections: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}

// UPDATED: Now uses composite partition key table for better performance
func (impl *collectionRepositoryImpl) FindByParent(ctx context.Context, parentID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	// Still use the original table for parent-child relationships since we need to query across all owners
	query := `SELECT collection_id FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id
		WHERE parent_id = ? AND state = ?`

	iter := impl.Session.Query(query, parentID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to find collections by parent: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}

// UPDATED: Now uses composite partition key for optimal performance - NO MORE ALLOW FILTERING!
func (impl *collectionRepositoryImpl) FindRootCollections(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID

	// Use the new composite partition key table for root collections
	nullParentID := impl.nullParentUUID()

	query := `SELECT collection_id FROM maplefile_collections_by_parent_and_owner_id_with_asc_created_at_and_asc_collection_id
		WHERE parent_id = ? AND owner_id = ? AND state = ?`

	iter := impl.Session.Query(query, nullParentID, ownerID, dom_collection.CollectionStateActive).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to find root collections: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}

// OPTIMIZED: No more recursive queries - single efficient query
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

	return impl.loadMultipleCollectionsWithMembers(ctx, descendantIDs)
}
