// cloud/mapleapps-backend/internal/maplefile/repo/collection/update.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
)

func (impl *collectionRepositoryImpl) Update(ctx context.Context, collection *dom_collection.Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	if !impl.isValidUUID(collection.ID) {
		return fmt.Errorf("collection ID is required")
	}

	// Get existing collection to compare changes
	existing, err := impl.GetWithAnyState(ctx, collection.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing collection: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("collection not found")
	}

	// Update modified timestamp
	collection.ModifiedAt = time.Now()

	// Serialize complex fields
	ancestorIDsJSON, err := impl.serializeAncestorIDs(collection.AncestorIDs)
	if err != nil {
		return fmt.Errorf("failed to serialize ancestor IDs: %w", err)
	}

	encryptedKeyJSON, err := impl.serializeEncryptedCollectionKey(collection.EncryptedCollectionKey)
	if err != nil {
		return fmt.Errorf("failed to serialize encrypted collection key: %w", err)
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Update main table
	batch.Query(`UPDATE maplefile_collections_by_id SET
		owner_id = ?, encrypted_name = ?, collection_type = ?, encrypted_collection_key = ?,
		parent_id = ?, ancestor_ids = ?, created_at = ?, created_by_user_id = ?,
		modified_at = ?, modified_by_user_id = ?, version = ?, state = ?,
		tombstone_version = ?, tombstone_expiry = ?
		WHERE id = ?`,
		collection.OwnerID, collection.EncryptedName, collection.CollectionType, encryptedKeyJSON,
		collection.ParentID, ancestorIDsJSON, collection.CreatedAt, collection.CreatedByUserID,
		collection.ModifiedAt, collection.ModifiedByUserID, collection.Version, collection.State,
		collection.TombstoneVersion, collection.TombstoneExpiry, collection.ID)

	// 2. Update user access - delete old owner entry and insert new one
	batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
		existing.OwnerID, existing.ModifiedAt, collection.ID)

	batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		(user_id, modified_at, collection_id, access_type, permission_level, state)
		VALUES (?, ?, ?, 'owner', ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, nil, collection.State)

	// 3. Update original parent index if parent changed
	oldParentID := existing.ParentID
	if !impl.isValidUUID(oldParentID) {
		oldParentID = impl.nullParentUUID()
	}

	newParentID := collection.ParentID
	if !impl.isValidUUID(newParentID) {
		newParentID = impl.nullParentUUID()
	}

	if oldParentID != newParentID || existing.OwnerID != collection.OwnerID {
		// Remove from old parent in original table
		batch.Query(`DELETE FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id
			WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
			oldParentID, collection.CreatedAt, collection.ID)

		// Add to new parent in original table
		batch.Query(`INSERT INTO maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id
			(parent_id, created_at, collection_id, owner_id, state)
			VALUES (?, ?, ?, ?, ?)`,
			newParentID, collection.CreatedAt, collection.ID, collection.OwnerID, collection.State)

		// Remove from old parent+owner in composite table
		batch.Query(`DELETE FROM maplefile_collections_by_parent_and_owner_id_with_asc_created_at_and_asc_collection_id
			WHERE parent_id = ? AND owner_id = ? AND created_at = ? AND collection_id = ?`,
			oldParentID, existing.OwnerID, collection.CreatedAt, collection.ID)

		// Add to new parent+owner in composite table
		batch.Query(`INSERT INTO maplefile_collections_by_parent_and_owner_id_with_asc_created_at_and_asc_collection_id
			(parent_id, owner_id, created_at, collection_id, state)
			VALUES (?, ?, ?, ?, ?)`,
			newParentID, collection.OwnerID, collection.CreatedAt, collection.ID, collection.State)
	} else {
		// Update existing parent entry in original table
		batch.Query(`UPDATE maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id SET
			owner_id = ?, state = ?
			WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
			collection.OwnerID, collection.State,
			newParentID, collection.CreatedAt, collection.ID)

		// Update existing parent entry in composite table
		batch.Query(`UPDATE maplefile_collections_by_parent_and_owner_id_with_asc_created_at_and_asc_collection_id SET
			state = ?
			WHERE parent_id = ? AND owner_id = ? AND created_at = ? AND collection_id = ?`,
			collection.State,
			newParentID, collection.OwnerID, collection.CreatedAt, collection.ID)
	}

	// 4. Update ancestor hierarchy
	oldAncestorEntries := impl.buildAncestorDepthEntries(collection.ID, existing.AncestorIDs)
	for _, entry := range oldAncestorEntries {
		batch.Query(`DELETE FROM maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			WHERE ancestor_id = ? AND depth = ? AND collection_id = ?`,
			entry.AncestorID, entry.Depth, entry.CollectionID)
	}

	newAncestorEntries := impl.buildAncestorDepthEntries(collection.ID, collection.AncestorIDs)
	for _, entry := range newAncestorEntries {
		batch.Query(`INSERT INTO maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			(ancestor_id, depth, collection_id, state)
			VALUES (?, ?, ?, ?)`,
			entry.AncestorID, entry.Depth, entry.CollectionID, collection.State)
	}

	// 5. Replace all members (delete old, insert new)
	batch.Query(`DELETE FROM maplefile_collection_members_by_collection_id_and_recipient_id WHERE collection_id = ?`, collection.ID)

	// Delete old member access entries (need to delete by individual user and timestamp)
	for _, oldMember := range existing.Members {
		batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
			oldMember.RecipientID, existing.ModifiedAt, collection.ID)
	}

	// Insert new members
	for _, member := range collection.Members {
		// Ensure member has an ID
		if !impl.isValidUUID(member.ID) {
			member.ID = gocql.TimeUUID()
		}

		batch.Query(`INSERT INTO maplefile_collection_members_by_collection_id_and_recipient_id
			(collection_id, recipient_id, member_id, recipient_email, granted_by_id,
			 encrypted_collection_key, permission_level, created_at,
			 is_inherited, inherited_from_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			collection.ID, member.RecipientID, member.ID, member.RecipientEmail,
			member.GrantedByID, member.EncryptedCollectionKey,
			member.PermissionLevel, member.CreatedAt,
			member.IsInherited, member.InheritedFromID)

		batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			(user_id, modified_at, collection_id, access_type, permission_level, state)
			VALUES (?, ?, ?, 'member', ?, ?)`,
			member.RecipientID, collection.ModifiedAt, collection.ID, member.PermissionLevel, collection.State)
	}

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch.WithContext(ctx)); err != nil {
		impl.Logger.Error("failed to update collection",
			zap.String("collection_id", collection.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update collection: %w", err)
	}

	impl.Logger.Info("collection updated successfully",
		zap.String("collection_id", collection.ID.String()))

	return nil
}
