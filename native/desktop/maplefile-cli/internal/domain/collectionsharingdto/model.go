// internal/domain/collection/model.go
package collectionsharingdto

import (
	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// ShareCollectionRequestDTO represents a request to share a collection
type ShareCollectionRequestDTO struct {
	CollectionID           gocql.UUID                   `json:"collection_id"`
	RecipientID            gocql.UUID                   `json:"recipient_id"`
	RecipientEmail         string                       `json:"recipient_email"`
	PermissionLevel        string                       `json:"permission_level"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
	ShareWithDescendants   bool                         `json:"share_with_descendants"`
}

// ShareCollectionResponseDTO represents the response from sharing a collection
type ShareCollectionResponseDTO struct {
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	MembershipsCreated int    `json:"memberships_created"`
}

// RemoveMemberRequest represents a request to remove a member from a collection
type RemoveMemberRequestDTO struct {
	CollectionID          gocql.UUID `json:"collection_id"`
	RecipientID           gocql.UUID `json:"recipient_id"`
	RemoveFromDescendants bool       `json:"remove_from_descendants"`
}

// RemoveMemberResponseDTO represents the response from removing a member
type RemoveMemberResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
