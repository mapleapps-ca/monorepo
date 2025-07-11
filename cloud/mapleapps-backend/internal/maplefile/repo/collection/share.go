// cloud/mapleapps-backend/internal/maplefile/repo/collection/share.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
)

func (impl *collectionRepositoryImpl) AddMember(ctx context.Context, collectionID gocql.UUID, membership *dom_collection.CollectionMembership) error {
	if membership == nil {
		return fmt.Errorf("membership cannot be nil")
	}

	impl.Logger.Info("starting add member process",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.String("recipient_email", membership.RecipientEmail),
		zap.String("permission_level", membership.PermissionLevel))

	// Validate membership data with enhanced checks
	if !impl.isValidUUID(membership.RecipientID) {
		return fmt.Errorf("invalid recipient ID")
	}
	if membership.RecipientEmail == "" {
		return fmt.Errorf("recipient email is required")
	}
	if membership.PermissionLevel == "" {
		membership.PermissionLevel = dom_collection.CollectionPermissionReadOnly
	}

	// CRITICAL: Validate encrypted collection key for shared members
	if len(membership.EncryptedCollectionKey) == 0 {
		impl.Logger.Error("CRITICAL: Attempt to add member without encrypted collection key",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.String("recipient_email", membership.RecipientEmail),
			zap.Int("encrypted_key_length", len(membership.EncryptedCollectionKey)))
		return fmt.Errorf("encrypted collection key is required for shared members")
	}

	// Additional validation: ensure the encrypted key is reasonable size
	if len(membership.EncryptedCollectionKey) < 32 {
		impl.Logger.Error("encrypted collection key appears too short",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Int("encrypted_key_length", len(membership.EncryptedCollectionKey)))
		return fmt.Errorf("encrypted collection key appears invalid (got %d bytes, expected at least 32)", len(membership.EncryptedCollectionKey))
	}

	impl.Logger.Info("validated encrypted collection key for new member",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.Int("encrypted_key_length", len(membership.EncryptedCollectionKey)))

	// Load collection
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		impl.Logger.Error("failed to get collection for member addition",
			zap.String("collection_id", collectionID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	impl.Logger.Info("loaded collection for member addition",
		zap.String("collection_id", collection.ID.String()),
		zap.String("collection_state", collection.State),
		zap.Int("existing_members", len(collection.Members)))

	// Ensure member has an ID BEFORE adding to collection
	if !impl.isValidUUID(membership.ID) {
		membership.ID = gocql.TimeUUID()
		impl.Logger.Debug("generated new member ID", zap.String("member_id", membership.ID.String()))
	}

	// Set creation time if not set
	if membership.CreatedAt.IsZero() {
		membership.CreatedAt = time.Now()
	}

	// Set collection ID (ensure it matches)
	membership.CollectionID = collectionID

	// Check if member already exists and update or add
	memberExists := false
	for i, existingMember := range collection.Members {
		if existingMember.RecipientID == membership.RecipientID {
			impl.Logger.Info("updating existing collection member",
				zap.String("collection_id", collectionID.String()),
				zap.String("recipient_id", membership.RecipientID.String()),
				zap.String("old_permission", existingMember.PermissionLevel),
				zap.String("new_permission", membership.PermissionLevel))

			// IMPORTANT: Preserve the existing member ID to avoid creating a new one
			membership.ID = existingMember.ID
			collection.Members[i] = *membership
			memberExists = true
			break
		}
	}

	if !memberExists {
		impl.Logger.Info("adding new collection member",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.String("permission_level", membership.PermissionLevel))

		collection.Members = append(collection.Members, *membership)

		impl.Logger.Info("DEBUGGING: Member added to collection.Members slice",
			zap.String("collection_id", collectionID.String()),
			zap.String("new_member_id", membership.ID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Int("total_members_now", len(collection.Members)))
	}

	// Update version
	collection.Version++
	collection.ModifiedAt = time.Now()

	impl.Logger.Info("prepared collection for update with member",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("total_members", len(collection.Members)),
		zap.Uint64("version", collection.Version))

	// DEBUGGING: Log all members that will be sent to Update method
	impl.Logger.Info("DEBUGGING: About to call Update() with these members:")
	for debugIdx, debugMember := range collection.Members {
		isOwner := debugMember.RecipientID == collection.OwnerID
		impl.Logger.Info("DEBUGGING: Member in collection.Members slice",
			zap.Int("index", debugIdx),
			zap.String("member_id", debugMember.ID.String()),
			zap.String("recipient_id", debugMember.RecipientID.String()),
			zap.String("recipient_email", debugMember.RecipientEmail),
			zap.String("permission_level", debugMember.PermissionLevel),
			zap.Bool("is_owner", isOwner),
			zap.Int("encrypted_key_length", len(debugMember.EncryptedCollectionKey)))
	}

	// Log all members for debugging
	for i, member := range collection.Members {
		isOwner := member.RecipientID == collection.OwnerID
		impl.Logger.Debug("collection member details",
			zap.Int("member_index", i),
			zap.String("member_id", member.ID.String()),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.String("recipient_email", member.RecipientEmail),
			zap.String("permission_level", member.PermissionLevel),
			zap.Bool("is_inherited", member.IsInherited),
			zap.Bool("is_owner", isOwner),
			zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))
	}

	// Call update - the Update method itself is atomic and reliable
	err = impl.Update(ctx, collection)
	if err != nil {
		impl.Logger.Error("failed to update collection with new member",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update collection: %w", err)
	}

	impl.Logger.Info("successfully added member to collection",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.String("member_id", membership.ID.String()))

	// DEBUGGING: Test if we can query the members table directly
	impl.Logger.Info("DEBUGGING: Testing direct access to members table")
	err = impl.testMembersTableAccess(ctx, collectionID)
	if err != nil {
		impl.Logger.Error("DEBUGGING: Failed to access members table",
			zap.String("collection_id", collectionID.String()),
			zap.Error(err))
	} else {
		impl.Logger.Info("DEBUGGING: Members table access test successful",
			zap.String("collection_id", collectionID.String()))
	}

	return nil
}

// testDirectMemberInsert tests inserting directly into the members table (for debugging)
func (impl *collectionRepositoryImpl) testDirectMemberInsert(ctx context.Context, collectionID gocql.UUID, membership *dom_collection.CollectionMembership) error {
	impl.Logger.Info("DEBUGGING: Testing direct insert into members table",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()))

	query := `INSERT INTO maplefile_collection_members_by_collection_id_and_recipient_id
		(collection_id, recipient_id, member_id, recipient_email, granted_by_id,
		 encrypted_collection_key, permission_level, created_at,
		 is_inherited, inherited_from_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := impl.Session.Query(query,
		collectionID, membership.RecipientID, membership.ID, membership.RecipientEmail,
		membership.GrantedByID, membership.EncryptedCollectionKey,
		membership.PermissionLevel, membership.CreatedAt,
		membership.IsInherited, membership.InheritedFromID).WithContext(ctx).Exec()

	if err != nil {
		impl.Logger.Error("DEBUGGING: Direct insert failed",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Error(err))
		return fmt.Errorf("direct insert failed: %w", err)
	}

	impl.Logger.Info("DEBUGGING: Direct insert successful",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()))

	// Verify the insert worked
	var foundMemberID gocql.UUID
	verifyQuery := `SELECT member_id FROM maplefile_collection_members_by_collection_id_and_recipient_id
		WHERE collection_id = ? AND recipient_id = ?`

	err = impl.Session.Query(verifyQuery, collectionID, membership.RecipientID).WithContext(ctx).Scan(&foundMemberID)
	if err != nil {
		if err == gocql.ErrNotFound {
			impl.Logger.Error("DEBUGGING: Direct insert verification failed - member not found",
				zap.String("collection_id", collectionID.String()),
				zap.String("recipient_id", membership.RecipientID.String()))
			return fmt.Errorf("direct insert verification failed - member not found")
		}
		impl.Logger.Error("DEBUGGING: Direct insert verification error",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Error(err))
		return fmt.Errorf("verification query failed: %w", err)
	}

	impl.Logger.Info("DEBUGGING: Direct insert verification successful",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.String("found_member_id", foundMemberID.String()))

	return nil
}

// testMembersTableAccess verifies we can read from the members table
func (impl *collectionRepositoryImpl) testMembersTableAccess(ctx context.Context, collectionID gocql.UUID) error {
	query := `SELECT COUNT(*) FROM maplefile_collection_members_by_collection_id_and_recipient_id WHERE collection_id = ?`

	var count int
	err := impl.Session.Query(query, collectionID).WithContext(ctx).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query members table: %w", err)
	}

	impl.Logger.Info("DEBUGGING: Members table query successful",
		zap.String("collection_id", collectionID.String()),
		zap.Int("member_count", count))

	return nil
}

func (impl *collectionRepositoryImpl) RemoveMember(ctx context.Context, collectionID, recipientID gocql.UUID) error {
	// Load collection, remove member, and save
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Remove member from collection
	var updatedMembers []dom_collection.CollectionMembership
	found := false

	for _, member := range collection.Members {
		if member.RecipientID != recipientID {
			updatedMembers = append(updatedMembers, member)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("member not found in collection")
	}

	collection.Members = updatedMembers
	collection.Version++

	return impl.Update(ctx, collection)
}

func (impl *collectionRepositoryImpl) UpdateMemberPermission(ctx context.Context, collectionID, recipientID gocql.UUID, newPermission string) error {
	// Load collection, update member permission, and save
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Update member permission
	found := false
	for i, member := range collection.Members {
		if member.RecipientID == recipientID {
			collection.Members[i].PermissionLevel = newPermission
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("member not found in collection")
	}

	collection.Version++
	return impl.Update(ctx, collection)
}

func (impl *collectionRepositoryImpl) GetCollectionMembership(ctx context.Context, collectionID, recipientID gocql.UUID) (*dom_collection.CollectionMembership, error) {
	var membership dom_collection.CollectionMembership

	query := `SELECT recipient_id, member_id, recipient_email, granted_by_id,
		encrypted_collection_key, permission_level, created_at,
		is_inherited, inherited_from_id
		FROM maplefile_collection_members_by_collection_id_and_recipient_id
		WHERE collection_id = ? AND recipient_id = ?`

	err := impl.Session.Query(query, collectionID, recipientID).WithContext(ctx).Scan(
		&membership.RecipientID, &membership.ID, &membership.RecipientEmail, &membership.GrantedByID,
		&membership.EncryptedCollectionKey, &membership.PermissionLevel,
		&membership.CreatedAt, &membership.IsInherited, &membership.InheritedFromID)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	membership.CollectionID = collectionID

	return &membership, nil
}

func (impl *collectionRepositoryImpl) AddMemberToHierarchy(ctx context.Context, rootID gocql.UUID, membership *dom_collection.CollectionMembership) error {
	// Get all descendants of the root collection
	descendants, err := impl.FindDescendants(ctx, rootID)
	if err != nil {
		return fmt.Errorf("failed to find descendants: %w", err)
	}

	impl.Logger.Info("adding member to collection hierarchy",
		zap.String("root_collection_id", rootID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.Int("descendants_count", len(descendants)))

	// Add to root collection
	if err := impl.AddMember(ctx, rootID, membership); err != nil {
		return fmt.Errorf("failed to add member to root collection: %w", err)
	}

	// Add to all descendants with inherited flag
	inheritedMembership := *membership
	inheritedMembership.IsInherited = true
	inheritedMembership.InheritedFromID = rootID

	successCount := 0
	for _, descendant := range descendants {
		// Generate new ID for each inherited membership
		inheritedMembership.ID = gocql.TimeUUID()

		if err := impl.AddMember(ctx, descendant.ID, &inheritedMembership); err != nil {
			impl.Logger.Warn("failed to add inherited member to descendant",
				zap.String("descendant_id", descendant.ID.String()),
				zap.String("recipient_id", membership.RecipientID.String()),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	impl.Logger.Info("completed hierarchy member addition",
		zap.String("root_collection_id", rootID.String()),
		zap.String("recipient_id", membership.RecipientID.String()),
		zap.Int("total_descendants", len(descendants)),
		zap.Int("successful_additions", successCount))

	return nil
}

func (impl *collectionRepositoryImpl) RemoveMemberFromHierarchy(ctx context.Context, rootID, recipientID gocql.UUID) error {
	// Get all descendants of the root collection
	descendants, err := impl.FindDescendants(ctx, rootID)
	if err != nil {
		return fmt.Errorf("failed to find descendants: %w", err)
	}

	// Remove from root collection
	if err := impl.RemoveMember(ctx, rootID, recipientID); err != nil {
		return fmt.Errorf("failed to remove member from root collection: %w", err)
	}

	// Remove from all descendants where access was inherited from this root
	for _, descendant := range descendants {
		// Only remove if the membership was inherited from this root
		membership, err := impl.GetCollectionMembership(ctx, descendant.ID, recipientID)
		if err != nil {
			impl.Logger.Warn("failed to get membership for descendant",
				zap.String("descendant_id", descendant.ID.String()),
				zap.Error(err))
			continue
		}

		if membership != nil && membership.IsInherited && membership.InheritedFromID == rootID {
			if err := impl.RemoveMember(ctx, descendant.ID, recipientID); err != nil {
				impl.Logger.Warn("failed to remove inherited member from descendant",
					zap.String("descendant_id", descendant.ID.String()),
					zap.String("recipient_id", recipientID.String()),
					zap.Error(err))
			}
		}
	}

	return nil
}
