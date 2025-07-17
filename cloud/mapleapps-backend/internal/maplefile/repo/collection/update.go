// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/collection/update.go
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

	impl.Logger.Info("starting collection update",
		zap.String("collection_id", collection.ID.String()),
		zap.Uint64("version", collection.Version),
		zap.Int("members_count", len(collection.Members)))

	// Get existing collection to compare changes
	existing, err := impl.Get(ctx, collection.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing collection: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("collection not found")
	}

	impl.Logger.Debug("loaded existing collection for comparison",
		zap.String("collection_id", existing.ID.String()),
		zap.Uint64("existing_version", existing.Version),
		zap.Int("existing_members_count", len(existing.Members)))

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

	//
	// 1. Update main table
	//

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

	//
	// 2. Update BOTH user access tables for owner
	//

	// Delete old owner entry from BOTH tables
	batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
		existing.OwnerID, existing.ModifiedAt, collection.ID)

	batch.Query(`DELETE FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = 'owner' AND modified_at = ? AND collection_id = ?`,
		existing.OwnerID, existing.ModifiedAt, collection.ID)

	// Insert new owner entry into BOTH tables
	batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		(user_id, modified_at, collection_id, access_type, permission_level, state)
		VALUES (?, ?, ?, 'owner', ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, nil, collection.State)

	batch.Query(`INSERT INTO maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		(user_id, access_type, modified_at, collection_id, permission_level, state)
		VALUES (?, 'owner', ?, ?, ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, nil, collection.State)

	//
	// 3. Update parent hierarchy if changed
	//

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

	//
	// 4. Update ancestor hierarchy
	//

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

	//
	// 5. Handle members - FIXED: Delete members individually with composite key
	//

	impl.Logger.Info("processing member updates",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("old_members", len(existing.Members)),
		zap.Int("new_members", len(collection.Members)))

	// Delete each existing member individually from the members table
	impl.Logger.Info("DEBUGGING: Deleting existing members individually from members table",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("existing_members_count", len(existing.Members)))

	for _, oldMember := range existing.Members {
		impl.Logger.Debug("deleting member from members table",
			zap.String("collection_id", collection.ID.String()),
			zap.String("recipient_id", oldMember.RecipientID.String()))

		batch.Query(`DELETE FROM maplefile_collection_members_by_collection_id_and_recipient_id
			WHERE collection_id = ? AND recipient_id = ?`,
			collection.ID, oldMember.RecipientID)
	}

	// Delete old member access entries from BOTH user access tables
	for _, oldMember := range existing.Members {
		impl.Logger.Debug("deleting old member access",
			zap.String("collection_id", collection.ID.String()),
			zap.String("recipient_id", oldMember.RecipientID.String()),
			zap.Time("old_modified_at", existing.ModifiedAt))

		// Delete from original table
		batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
			oldMember.RecipientID, existing.ModifiedAt, collection.ID)

		// Delete from access-type-specific table
		batch.Query(`DELETE FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'member' AND modified_at = ? AND collection_id = ?`,
			oldMember.RecipientID, existing.ModifiedAt, collection.ID)
	}

	// Insert ALL new members into ALL tables
	impl.Logger.Info("DEBUGGING: About to insert members into tables",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("total_members_to_insert", len(collection.Members)))

	for i, member := range collection.Members {
		impl.Logger.Info("inserting new member",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("member_index", i),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.String("recipient_email", member.RecipientEmail),
			zap.String("permission_level", member.PermissionLevel),
			zap.Bool("is_inherited", member.IsInherited))

		// Validate member data before insertion
		if !impl.isValidUUID(member.RecipientID) {
			return fmt.Errorf("invalid recipient ID for member %d", i)
		}
		if member.RecipientEmail == "" {
			return fmt.Errorf("recipient email is required for member %d", i)
		}
		if member.PermissionLevel == "" {
			return fmt.Errorf("permission level is required for member %d", i)
		}

		// FIXED: Only require encrypted collection key for non-owner members
		// The owner has access to the collection key through their master key
		isOwner := member.RecipientID == collection.OwnerID
		if !isOwner && len(member.EncryptedCollectionKey) == 0 {
			impl.Logger.Error("CRITICAL: encrypted collection key missing for shared member",
				zap.String("collection_id", collection.ID.String()),
				zap.Int("member_index", i),
				zap.String("recipient_id", member.RecipientID.String()),
				zap.String("recipient_email", member.RecipientEmail),
				zap.String("owner_id", collection.OwnerID.String()),
				zap.Bool("is_owner", isOwner),
				zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))
			return fmt.Errorf("VALIDATION ERROR: encrypted collection key is required for shared member %d (recipient: %s, email: %s). This indicates a frontend bug or API misuse.", i, member.RecipientID.String(), member.RecipientEmail)
		}

		// Additional validation for shared members
		if !isOwner && len(member.EncryptedCollectionKey) > 0 && len(member.EncryptedCollectionKey) < 32 {
			impl.Logger.Error("encrypted collection key appears invalid for shared member",
				zap.String("collection_id", collection.ID.String()),
				zap.Int("member_index", i),
				zap.String("recipient_id", member.RecipientID.String()),
				zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))
			return fmt.Errorf("encrypted collection key appears invalid for member %d (too short: %d bytes)", i, len(member.EncryptedCollectionKey))
		}

		// Log key status for debugging
		impl.Logger.Debug("member key validation passed",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("member_index", i),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.Bool("is_owner", isOwner),
			zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))

		// Ensure member has an ID - but don't regenerate if it already exists
		if !impl.isValidUUID(member.ID) {
			member.ID = gocql.TimeUUID()
			impl.Logger.Debug("generated member ID",
				zap.String("member_id", member.ID.String()),
				zap.String("recipient_id", member.RecipientID.String()))
		} else {
			impl.Logger.Debug("using existing member ID",
				zap.String("member_id", member.ID.String()),
				zap.String("recipient_id", member.RecipientID.String()))
		}

		// Insert into normalized members table
		impl.Logger.Info("DEBUGGING: Inserting member into members table",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("member_index", i),
			zap.String("member_id", member.ID.String()),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.String("recipient_email", member.RecipientEmail),
			zap.String("permission_level", member.PermissionLevel))

		batch.Query(`INSERT INTO maplefile_collection_members_by_collection_id_and_recipient_id
			(collection_id, recipient_id, member_id, recipient_email, granted_by_id,
			 encrypted_collection_key, permission_level, created_at,
			 is_inherited, inherited_from_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			collection.ID, member.RecipientID, member.ID, member.RecipientEmail,
			member.GrantedByID, member.EncryptedCollectionKey,
			member.PermissionLevel, member.CreatedAt,
			member.IsInherited, member.InheritedFromID)

		impl.Logger.Info("DEBUGGING: Added member insert query to batch",
			zap.String("collection_id", collection.ID.String()),
			zap.String("member_id", member.ID.String()),
			zap.String("recipient_id", member.RecipientID.String()))

		// Insert into BOTH user access tables
		// Original table
		batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			(user_id, modified_at, collection_id, access_type, permission_level, state)
			VALUES (?, ?, ?, 'member', ?, ?)`,
			member.RecipientID, collection.ModifiedAt, collection.ID, member.PermissionLevel, collection.State)

		// Access-type-specific table
		batch.Query(`INSERT INTO maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			(user_id, access_type, modified_at, collection_id, permission_level, state)
			VALUES (?, 'member', ?, ?, ?, ?)`,
			member.RecipientID, collection.ModifiedAt, collection.ID, member.PermissionLevel, collection.State)
	}

	//
	// 6. Execute the batch
	//

	impl.Logger.Info("executing batch update",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("batch_size", batch.Size()))

	// Execute batch - ensures atomicity across all table updates
	impl.Logger.Info("DEBUGGING: About to execute batch with member inserts",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("batch_size", batch.Size()),
		zap.Int("members_in_batch", len(collection.Members)))

	if err := impl.Session.ExecuteBatch(batch.WithContext(ctx)); err != nil {
		impl.Logger.Error("DEBUGGING: Batch execution failed",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("batch_size", batch.Size()),
			zap.Error(err))
		return fmt.Errorf("failed to update collection: %w", err)
	}

	impl.Logger.Info("DEBUGGING: Batch execution completed successfully",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("batch_size", batch.Size()))

	// Remove the immediate verification - Cassandra needs time to propagate
	// In production, we should trust the batch succeeded if no error was returned

	impl.Logger.Info("collection updated successfully in all tables",
		zap.String("collection_id", collection.ID.String()),
		zap.String("old_owner", existing.OwnerID.String()),
		zap.String("new_owner", collection.OwnerID.String()),
		zap.Int("old_member_count", len(existing.Members)),
		zap.Int("new_member_count", len(collection.Members)))

	return nil
}
