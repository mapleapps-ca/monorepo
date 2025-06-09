// cloud/backend/internal/maplefile/repo/collection/create.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

// cloud/mapleapps-backend/internal/maplefile/repo/collection/create.go
func (impl *collectionRepositoryImpl) Create(ctx context.Context, collection *dom_collection.Collection) error {
	// Validate collection
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

	// Use batch for consistency across multiple tables
	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Insert into main table
	batch.Query(`INSERT INTO maplefile_collections_by_id
		(id, owner_id, encrypted_name, collection_type, encrypted_collection_key,
		 members, parent_id, ancestor_ids, created_at, created_by_user_id,
		 modified_at, modified_by_user_id, version, state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		collection.ID, collection.OwnerID, collection.EncryptedName, collection.CollectionType,
		encryptedKeyJSON, membersJSON, collection.ParentID, ancestorIDsJSON,
		collection.CreatedAt, collection.CreatedByUserID, collection.ModifiedAt,
		collection.ModifiedByUserID, collection.Version, collection.State,
		collection.TombstoneVersion, collection.TombstoneExpiry)

	// 2. Insert into owner index
	batch.Query(`INSERT INTO maplefile_collections_by_owner_id_with_desc_modified_at_and_asc_id
		(owner_id, modified_at, collection_id, state, parent_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		collection.OwnerID, collection.ModifiedAt, collection.ID, collection.State,
		collection.ParentID, collection.CreatedAt)

	// 3. Insert into parent index (if has parent)
	if impl.isValidUUID(collection.ParentID) {
		batch.Query(`INSERT INTO maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
			(parent_id, created_at, collection_id, state)
			VALUES (?, ?, ?, ?)`,
			collection.ParentID, collection.CreatedAt, collection.ID, collection.State)
	}

	// 4. Insert into ancestor depth index
	ancestorEntries := impl.buildAncestorDepthEntries(collection.ID, collection.AncestorIDs)
	for _, entry := range ancestorEntries {
		batch.Query(`INSERT INTO maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			(ancestor_id, depth, collection_id, state)
			VALUES (?, ?, ?, ?)`,
			entry.AncestorID, entry.Depth, entry.CollectionID, collection.State)
	}

	// 5. Insert membership entries
	for _, member := range collection.Members {
		batch.Query(`INSERT INTO maplefile_collections_by_recipient_id_and_collection_id
			(recipient_id, collection_id, permission_level, granted_at, modified_at)
			VALUES (?, ?, ?, ?, ?)`,
			member.RecipientID, collection.ID, member.PermissionLevel,
			member.CreatedAt, collection.ModifiedAt)
	}

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to create collection",
			zap.String("collection_id", collection.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create collection: %w", err)
	}

	impl.Logger.Info("collection created successfully",
		zap.String("collection_id", collection.ID.String()),
		zap.String("owner_id", collection.OwnerID.String()))

	return nil
}
