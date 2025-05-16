// monorepo/cloud/backend/internal/maplefile/domain/collection/model.go
package collection

import (
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
)

const (
	CollectionTypeFolder = "folder"
	CollectionTypeAlbum  = "album"

	// Permission levels
	CollectionPermissionReadOnly  = "read_only"
	CollectionPermissionReadWrite = "read_write"
	CollectionPermissionAdmin     = "admin"
)

// Collection represents a folder or album
type Collection struct {
	ID        string    `bson:"id" json:"id"`
	OwnerID   string    `bson:"owner_id" json:"owner_id"`
	Name      string    `bson:"name" json:"name"` // Encrypted
	Path      string    `bson:"path" json:"path"` // Encrypted
	Type      string    `bson:"type" json:"type"` // "folder" or "album"
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// Collection key encrypted with owner's master key
	EncryptedCollectionKey keys.EncryptedCollectionKey `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Collection members (users with access)
	Members []CollectionMembership `bson:"members" json:"members"`
}

// CollectionMembership represents a user's access to a collection
type CollectionMembership struct {
	CollectionID   string `bson:"collection_id" json:"collection_id"`     // ID of the collection (redundant but helpful for queries)
	RecipientID    string `bson:"recipient_id" json:"recipient_id"`       // User receiving access
	RecipientEmail string `bson:"recipient_email" json:"recipient_email"` // Email for display purposes
	GrantedByID    string `bson:"granted_by_id" json:"granted_by_id"`     // User who shared the collection

	// Collection key encrypted with recipient's public key using box_seal. This matches the box_seal format which doesn't need a separate nonce.
	EncryptedCollectionKey []byte `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	PermissionLevel string    `bson:"permission_level" json:"permission_level"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}
