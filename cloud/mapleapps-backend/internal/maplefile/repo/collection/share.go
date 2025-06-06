// cloud/backend/internal/maplefile/repo/collection/share.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

// AddMember adds a new member to a collection, giving them access
func (impl collectionRepositoryImpl) AddMember(ctx context.Context, collectionID gocql.UUID, membership *dom_collection.CollectionMembership) error {

	return nil
}

// RemoveMember removes a member from a collection, revoking their access
func (impl collectionRepositoryImpl) RemoveMember(ctx context.Context, collectionID, recipientID gocql.UUID) error {

	return nil
}

// UpdateMemberPermission updates a member's permission level
func (impl collectionRepositoryImpl) UpdateMemberPermission(ctx context.Context, collectionID, recipientID gocql.UUID, newPermission string) error {

	return nil
}

// GetCollectionMembership retrieves a specific membership from a collection
func (impl collectionRepositoryImpl) GetCollectionMembership(ctx context.Context, collectionID, recipientID gocql.UUID) (*dom_collection.CollectionMembership, error) {

	return nil, nil
}

// AddMemberToHierarchy adds a user to a collection and all of its descendants
func (impl collectionRepositoryImpl) AddMemberToHierarchy(ctx context.Context, rootID gocql.UUID, membership *dom_collection.CollectionMembership) error {

	return nil
}

// RemoveMemberFromHierarchy removes a user from a collection and all of its descendants
func (impl collectionRepositoryImpl) RemoveMemberFromHierarchy(ctx context.Context, rootID, recipientID gocql.UUID) error {

	return nil
}
