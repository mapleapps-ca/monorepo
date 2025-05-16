// monorepo/cloud/backend/internal/maplefile/domain/collection/model.go
package collection

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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
// Can be used for both root collections and embedded subcollections
type Collection struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	OwnerID       primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	EncryptedName string             `bson:"encrypted_name" json:"encrypted_name"`
	Type          string             `bson:"type" json:"type"` // "folder" or "album"
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	ModifiedAt    time.Time          `bson:"modified_at" json:"modified_at"`

	// Collection key encrypted with owner's master key
	EncryptedCollectionKey keys.EncryptedCollectionKey `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Collection members (users with access)
	Members []CollectionMembership `bson:"members" json:"members"`

	// Hierarchical structure fields
	ParentID    primitive.ObjectID   `bson:"parent_id,omitempty" json:"parent_id,omitempty"`       // Parent collection ID, not stored for root collections
	AncestorIDs []primitive.ObjectID `bson:"ancestor_ids,omitempty" json:"ancestor_ids,omitempty"` // Array of ancestor IDs for efficient querying

	// EncryptedPathSegments information for efficient querying
	// Stores the path from root to this collection
	EncryptedPathSegments []string `bson:"encrypted_path_segments" json:"encrypted_path_segments"`

	// Child collections embedded directly within parent
	// Recursive embedding of the same type
	Children []*Collection `bson:"children,omitempty" json:"children,omitempty"`
}

// CollectionMembership represents a user's access to a collection
type CollectionMembership struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	CollectionID   primitive.ObjectID `bson:"collection_id" json:"collection_id"`     // ID of the collection (redundant but helpful for queries)
	RecipientID    primitive.ObjectID `bson:"recipient_id" json:"recipient_id"`       // User receiving access
	RecipientEmail string             `bson:"recipient_email" json:"recipient_email"` // Email for display purposes
	GrantedByID    primitive.ObjectID `bson:"granted_by_id" json:"granted_by_id"`     // User who shared the collection

	// Collection key encrypted with recipient's public key using box_seal. This matches the box_seal format which doesn't need a separate nonce.
	EncryptedCollectionKey []byte `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Access details
	PermissionLevel string    `bson:"permission_level" json:"permission_level"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`

	// Sharing origin tracking
	IsInherited     bool               `bson:"is_inherited" json:"is_inherited"`                               // Tracks whether access was granted directly or inherited from a parent
	InheritedFromID primitive.ObjectID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"` // InheritedFromID identifies which parent collection granted this access
}
