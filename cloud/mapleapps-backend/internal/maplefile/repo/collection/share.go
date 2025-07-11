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

	// Validate membership data
	if !impl.isValidUUID(membership.RecipientID) {
		return fmt.Errorf("invalid recipient ID")
	}
	if membership.RecipientEmail == "" {
		return fmt.Errorf("recipient email is required")
	}
	if membership.PermissionLevel == "" {
		membership.PermissionLevel = dom_collection.CollectionPermissionReadOnly
	}
	if len(membership.EncryptedCollectionKey) == 0 {
		return fmt.Errorf("encrypted collection key is required")
	}

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

	// Ensure member has an ID
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
	}

	// Update version
	collection.Version++
	collection.ModifiedAt = time.Now()

	impl.Logger.Info("prepared collection for update with member",
		zap.String("collection_id", collection.ID.String()),
		zap.Int("total_members", len(collection.Members)),
		zap.Uint64("version", collection.Version))

	// Log all members for debugging
	for i, member := range collection.Members {
		impl.Logger.Debug("collection member details",
			zap.Int("member_index", i),
			zap.String("member_id", member.ID.String()),
			zap.String("recipient_id", member.RecipientID.String()),
			zap.String("recipient_email", member.RecipientEmail),
			zap.String("permission_level", member.PermissionLevel),
			zap.Bool("is_inherited", member.IsInherited),
			zap.Int("encrypted_key_length", len(member.EncryptedCollectionKey)))
	}

	// Call update
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
		zap.String("recipient_id", membership.RecipientID.String()))

	// Verify the member was actually added by querying the members table
	err = impl.verifyMembershipCreated(ctx, collectionID, membership.RecipientID)
	if err != nil {
		impl.Logger.Error("member verification failed",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", membership.RecipientID.String()),
			zap.Error(err))
		return fmt.Errorf("member addition verification failed: %w", err)
	}

	return nil
}

// Helper method to verify membership was created in the database
func (impl *collectionRepositoryImpl) verifyMembershipCreated(ctx context.Context, collectionID, recipientID gocql.UUID) error {
	var memberID gocql.UUID
	query := `SELECT member_id FROM maplefile_collection_members_by_collection_id_and_recipient_id
		WHERE collection_id = ? AND recipient_id = ?`

	err := impl.Session.Query(query, collectionID, recipientID).WithContext(ctx).Scan(&memberID)
	if err != nil {
		if err == gocql.ErrNotFound {
			return fmt.Errorf("member not found in members table after creation")
		}
		return fmt.Errorf("failed to verify member creation: %w", err)
	}

	impl.Logger.Info("verified member exists in members table",
		zap.String("collection_id", collectionID.String()),
		zap.String("recipient_id", recipientID.String()),
		zap.String("member_id", memberID.String()))

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
