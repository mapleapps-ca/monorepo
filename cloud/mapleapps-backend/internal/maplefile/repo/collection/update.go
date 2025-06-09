// cloud/backend/internal/maplefile/repo/collection/update.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
)

// cloud/mapleapps-backend/internal/maplefile/repo/collection/update.go
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
	membersJSON, err := impl.serializeMembers(collection.Members)
	if err != nil {
		return fmt.Errorf("failed to serialize members: %w", err)
	}

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
		encrypted_name = ?, collection_type = ?, encrypted_collection_key = ?,
		members = ?, parent_id = ?, ancestor_ids = ?, modified_at = ?,
		modified_by_user_id = ?, version = ?, state = ?,
		tombstone_version = ?, tombstone_expiry = ?
		WHERE id = ?`,
		collection.EncryptedName, collection.CollectionType, encryptedKeyJSON,
		membersJSON, collection.ParentID, ancestorIDsJSON, collection.ModifiedAt,
		collection.ModifiedByUserID, collection.Version, collection.State,
		collection.TombstoneVersion, collection.TombstoneExpiry, collection.ID)

	// 2. Update owner index (modified_at changed)
	batch.Query(`UPDATE maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id SET
		state = ?, parent_id = ? WHERE owner_id = ? AND modified_at = ? AND collection_id = ?`,
		collection.State, collection.ParentID, collection.OwnerID, existing.ModifiedAt, collection.ID)

	// Delete old entry and insert new one with new modified_at
	batch.Query(`DELETE FROM maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id
		WHERE owner_id = ? AND modified_at = ? AND collection_id = ?`,
		collection.OwnerID, existing.ModifiedAt, collection.ID)

	batch.Query(`INSERT INTO maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id
		(owner_id, modified_at, collection_id, state, parent_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, collection.State,
		collection.ParentID, collection.CreatedAt)

	// 3. Update parent index if parent changed
	if existing.ParentID != collection.ParentID {
		// Remove from old parent
		if impl.isValidUUID(existing.ParentID) {
			batch.Query(`DELETE FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
				WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
				existing.ParentID, collection.CreatedAt, collection.ID)
		}

		// Add to new parent
		if impl.isValidUUID(collection.ParentID) {
			batch.Query(`INSERT INTO maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
				(parent_id, created_at, collection_id, state)
				VALUES (?, ?, ?, ?)`,
				collection.ParentID, collection.CreatedAt, collection.ID, collection.State)
		}
	} else if impl.isValidUUID(collection.ParentID) {
		// Update state in parent index
		batch.Query(`UPDATE maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id SET
			state = ? WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
			collection.State, collection.ParentID, collection.CreatedAt, collection.ID)
	}

	// 4. Update ancestor depth index (remove old, add new)
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

	// 5. Update membership index
	// Remove old memberships
	for _, oldMember := range existing.Members {
		batch.Query(`DELETE FROM maplefile_collections_by_recipient_id_and_collection_id
			WHERE recipient_id = ? AND collection_id = ?`,
			oldMember.RecipientID, collection.ID)
	}

	// Add new memberships
	for _, newMember := range collection.Members {
		batch.Query(`INSERT INTO maplefile_collections_by_recipient_id_and_collection_id
			(recipient_id, collection_id, permission_level, granted_at, modified_at)
			VALUES (?, ?, ?, ?, ?)`,
			newMember.RecipientID, collection.ID, newMember.PermissionLevel,
			newMember.CreatedAt, collection.ModifiedAt)
	}

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to update collection",
			zap.String("collection_id", collection.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update collection: %w", err)
	}

	impl.Logger.Info("collection updated successfully",
		zap.String("collection_id", collection.ID.String()))

	return nil
}
