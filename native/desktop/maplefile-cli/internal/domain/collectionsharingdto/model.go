// internal/domain/collection/model.go
package collectionsharingdto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShareCollectionRequestDTO represents a request to share a collection
type ShareCollectionRequestDTO struct {
	CollectionID           primitive.ObjectID `json:"collection_id"`
	RecipientID            primitive.ObjectID `json:"recipient_id"`
	RecipientEmail         string             `json:"recipient_email"`
	PermissionLevel        string             `json:"permission_level"`
	EncryptedCollectionKey string             `json:"encrypted_collection_key"`
	ShareWithDescendants   bool               `json:"share_with_descendants"`
}

// ShareCollectionResponseDTO represents the response from sharing a collection
type ShareCollectionResponseDTO struct {
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	MembershipsCreated int    `json:"memberships_created"`
}

// RemoveMemberRequest represents a request to remove a member from a collection
type RemoveMemberRequestDTO struct {
	CollectionID          primitive.ObjectID `json:"collection_id"`
	RecipientID           primitive.ObjectID `json:"recipient_id"`
	RemoveFromDescendants bool               `json:"remove_from_descendants"`
}

// RemoveMemberResponseDTO represents the response from removing a member
type RemoveMemberResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
