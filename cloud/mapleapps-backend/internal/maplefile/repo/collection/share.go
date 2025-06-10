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

	// Load collection, update members, and save
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Ensure member has an ID
	if !impl.isValidUUID(membership.ID) {
		membership.ID = gocql.TimeUUID()
	}

	// Set creation time if not set
	if membership.CreatedAt.IsZero() {
		membership.CreatedAt = time.Now()
	}

	// Check if member already exists
	for i, existingMember := range collection.Members {
		if existingMember.RecipientID == membership.RecipientID {
			// Update existing member
			collection.Members[i] = *membership
			collection.Version++
			return impl.Update(ctx, collection)
		}
	}

	// Add new member
	collection.Members = append(collection.Members, *membership)
	collection.Version++

	return impl.Update(ctx, collection)
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
		FROM maplefile_collection_members
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

	// Add to root collection
	if err := impl.AddMember(ctx, rootID, membership); err != nil {
		return fmt.Errorf("failed to add member to root collection: %w", err)
	}

	// Add to all descendants with inherited flag
	inheritedMembership := *membership
	inheritedMembership.IsInherited = true
	inheritedMembership.InheritedFromID = rootID

	for _, descendant := range descendants {
		// Generate new ID for each inherited membership
		inheritedMembership.ID = gocql.TimeUUID()

		if err := impl.AddMember(ctx, descendant.ID, &inheritedMembership); err != nil {
			impl.Logger.Warn("failed to add inherited member to descendant",
				zap.String("descendant_id", descendant.ID.String()),
				zap.String("recipient_id", membership.RecipientID.String()),
				zap.Error(err))
		}
	}

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
