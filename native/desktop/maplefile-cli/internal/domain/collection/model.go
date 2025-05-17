// internal/domain/collection/model.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// Collection represents a collection in the system
type Collection struct {
	// Existing cloud fields
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`

	// New fields for local decrypted data
	DecryptedName     string    `json:"-" bson:"-"`
	DecryptedPath     string    `json:"-" bson:"-"`
	LastSyncedAt      time.Time `json:"-" bson:"-"`
	IsModifiedLocally bool      `json:"-" bson:"-"`
}

// CreateCollectionRequest represents the data needed to create a collection
type CreateCollectionRequest struct {
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

// CollectionResponse represents the server's response when creating a collection
type CollectionResponse struct {
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`
	Members                []MembershipResponse        `json:"members"`
}

// MembershipResponse represents a collection membership
type MembershipResponse struct {
	ID              primitive.ObjectID `json:"id"`
	RecipientID     primitive.ObjectID `json:"recipient_id"`
	RecipientEmail  string             `json:"recipient_email"`
	PermissionLevel string             `json:"permission_level"`
	GrantedByID     primitive.ObjectID `json:"granted_by_id"`
	CollectionID    primitive.ObjectID `json:"collection_id"`
	IsInherited     bool               `json:"is_inherited"`
	InheritedFromID primitive.ObjectID `json:"inherited_from_id,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
}

// CollectionRepository defines the interface for interacting with collections
type CollectionRepository interface {
	CreateCollection(ctx context.Context, request *CreateCollectionRequest) (*CollectionResponse, error)
}

// Constants for collection types
const (
	CollectionTypeFolder = "folder"
	CollectionTypeAlbum  = "album"
)
