// cloud/mapleapps-backend/internal/maplefile/repo/collection/create.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) Create(ctx context.Context, collection *dom_collection.Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	if !impl.isValidUUID(collection.ID) {
		return fmt.Errorf("collection ID is required")
	}

	if !impl.isValidUUID(collection.OwnerID) {
		return fmt.Errorf("owner ID is required")
	}

	// Set creation timestamp if not set
	if collection.CreatedAt.IsZero() {
		collection.CreatedAt = time.Now()
	}

	if collection.ModifiedAt.IsZero() {
		collection.ModifiedAt = collection.CreatedAt
	}

	// Ensure state is set
	if collection.State == "" {
		collection.State = dom_collection.CollectionStateActive
	}

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

	// 1. Insert into main table
	batch.Query(`INSERT INTO maplefile_collections_by_id
		(id, owner_id, encrypted_name, collection_type, encrypted_collection_key,
		 parent_id, ancestor_ids, created_at, created_by_user_id,
		 modified_at, modified_by_user_id, version, state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		collection.ID, collection.OwnerID, collection.EncryptedName, collection.CollectionType,
		encryptedKeyJSON, collection.ParentID, ancestorIDsJSON,
		collection.CreatedAt, collection.CreatedByUserID, collection.ModifiedAt,
		collection.ModifiedByUserID, collection.Version, collection.State,
		collection.TombstoneVersion, collection.TombstoneExpiry)

	// 2. Insert owner access into BOTH user access tables

	// 2 -> (1 of 2): Original table: supports queries across all access types
	batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		(user_id, modified_at, collection_id, access_type, permission_level, state)
		VALUES (?, ?, ?, 'owner', ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, nil, collection.State)

	// 2 -> (2 of 2): Access-type-specific table for efficient filtering
	batch.Query(`INSERT INTO maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		(user_id, access_type, modified_at, collection_id, permission_level, state)
		VALUES (?, 'owner', ?, ?, ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, nil, collection.State)

	// 3. Insert into original parent index (still needed for cross-owner parent-child queries)
	parentID := collection.ParentID
	if !impl.isValidUUID(parentID) {
		parentID = impl.nullParentUUID() // Use null UUID for root collections
	}

	batch.Query(`INSERT INTO maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id
		(parent_id, created_at, collection_id, owner_id, state)
		VALUES (?, ?, ?, ?, ?)`,
		parentID, collection.CreatedAt, collection.ID, collection.OwnerID, collection.State)

	// 4. Insert into composite partition key table for optimized root collection queries
	batch.Query(`INSERT INTO maplefile_collections_by_parent_and_owner_id_with_asc_created_at_and_asc_collection_id
		(parent_id, owner_id, created_at, collection_id, state)
		VALUES (?, ?, ?, ?, ?)`,
		parentID, collection.OwnerID, collection.CreatedAt, collection.ID, collection.State)

	// 5. Insert into ancestor hierarchy table
	ancestorEntries := impl.buildAncestorDepthEntries(collection.ID, collection.AncestorIDs)
	for _, entry := range ancestorEntries {
		batch.Query(`INSERT INTO maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			(ancestor_id, depth, collection_id, state)
			VALUES (?, ?, ?, ?)`,
			entry.AncestorID, entry.Depth, entry.CollectionID, collection.State)
	}

	// 6. Insert members into normalized table AND both user access tables - WITH CONSISTENT VALIDATION
	for i, member := range collection.Members {
		impl.Logger.Info("processing member for creation",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("member_index", i),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.String("recipient_email", member.RecipientEmail),
			zap.String("permission_level", member.PermissionLevel),
			zap.Bool("is_inherited", member.IsInherited))

		// Validate member data before insertion - CONSISTENT WITH UPDATE METHOD
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
			impl.Logger.Error("CRITICAL: encrypted collection key missing for shared member during creation",
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
			impl.Logger.Error("encrypted collection key appears invalid for shared member during creation",
				zap.String("collection_id", collection.ID.String()),
				zap.Int("member_index", i),
				zap.String("recipient_id", member.RecipientID.String()),
				zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))
			return fmt.Errorf("encrypted collection key appears invalid for member %d (too short: %d bytes)", i, len(member.EncryptedCollectionKey))
		}

		// Log key status for debugging
		impl.Logger.Debug("member key validation passed during creation",
			zap.String("collection_id", collection.ID.String()),
			zap.Int("member_index", i),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.Bool("is_owner", isOwner),
			zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))

		// Ensure member has an ID - but don't regenerate if it already exists
		if !impl.isValidUUID(member.ID) {
			member.ID = gocql.TimeUUID()
			impl.Logger.Debug("generated member ID during creation",
				zap.String("member_id", member.ID.String()),
				zap.String("recipient_id", member.RecipientID.String()))
		} else {
			impl.Logger.Debug("using existing member ID during creation",
				zap.String("member_id", member.ID.String()),
				zap.String("recipient_id", member.RecipientID.String()))
		}

		// Insert into normalized members table
		batch.Query(`INSERT INTO maplefile_collection_members_by_collection_id_and_recipient_id
			(collection_id, recipient_id, member_id, recipient_email, granted_by_id,
			 encrypted_collection_key, permission_level, created_at,
			 is_inherited, inherited_from_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			collection.ID, member.RecipientID, member.ID, member.RecipientEmail,
			member.GrantedByID, member.EncryptedCollectionKey,
			member.PermissionLevel, member.CreatedAt,
			member.IsInherited, member.InheritedFromID)

		// Add member access to BOTH user access tables
		// Original table: supports all-access-types queries
		batch.Query(`INSERT INTO maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			(user_id, modified_at, collection_id, access_type, permission_level, state)
			VALUES (?, ?, ?, 'member', ?, ?)`,
			member.RecipientID, collection.ModifiedAt, collection.ID, member.PermissionLevel, collection.State)

		// NEW: Access-type-specific table for efficient member queries
		batch.Query(`INSERT INTO maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			(user_id, access_type, modified_at, collection_id, permission_level, state)
			VALUES (?, 'member', ?, ?, ?, ?)`,
			member.RecipientID, collection.ModifiedAt, collection.ID, member.PermissionLevel, collection.State)
	}

	// Execute batch - this ensures all tables are updated atomically
	if err := impl.Session.ExecuteBatch(batch.WithContext(ctx)); err != nil {
		impl.Logger.Error("failed to create collection",
			zap.String("collection_id", collection.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create collection: %w", err)
	}

	impl.Logger.Info("collection created successfully in all tables",
		zap.String("collection_id", collection.ID.String()),
		zap.String("owner_id", collection.OwnerID.String()),
		zap.Int("member_count", len(collection.Members)))

	return nil
}
