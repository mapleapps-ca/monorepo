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

	// 1. Update main collection table with ALL Collection struct fields (except Members)
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

	// 2. Update owner access entry
	batch.Query(`DELETE FROM maplefile_collections_by_user
		WHERE user_id = ? AND collection_id = ? AND access_type = 'owner'`,
		collection.OwnerID, collection.ID)

	batch.Query(`INSERT INTO maplefile_collections_by_user
		(user_id, collection_id, access_type, modified_at, state, parent_id, created_at)
		VALUES (?, ?, 'owner', ?, ?, ?, ?)`,
		collection.OwnerID, collection.ID, collection.ModifiedAt,
		collection.State, collection.ParentID, collection.CreatedAt)

	// 3. Replace all members (delete old, insert new)
	batch.Query(`DELETE FROM maplefile_collection_members WHERE collection_id = ?`, collection.ID)

	batch.Query(`DELETE FROM maplefile_collections_by_user
		WHERE collection_id = ? AND access_type = 'member'`, collection.ID)

	// 4. Insert updated members with all CollectionMembership fields
	for _, member := range collection.Members {
		batch.Query(`INSERT INTO maplefile_collection_members
			(collection_id, recipient_id, member_id, recipient_email, granted_by_id,
			 encrypted_collection_key, permission_level, created_at,
			 is_inherited, inherited_from_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			collection.ID, member.RecipientID, member.ID, member.RecipientEmail,
			member.GrantedByID, member.EncryptedCollectionKey,
			member.PermissionLevel, member.CreatedAt,
			member.IsInherited, member.InheritedFromID)

		batch.Query(`INSERT INTO maplefile_collections_by_user
			(user_id, collection_id, access_type, permission_level, modified_at, state, parent_id, created_at)
			VALUES (?, ?, 'member', ?, ?, ?, ?, ?)`,
			member.RecipientID, collection.ID, member.PermissionLevel,
			collection.ModifiedAt, collection.State, collection.ParentID, collection.CreatedAt)
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
