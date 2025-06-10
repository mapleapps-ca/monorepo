// monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto/model.go
package collectiondto

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// CollectionDTO represents a Data Transfer Object (DTO) used for transferring an encrypted collection entity
// to and from a cloud service. This entity holds metadata about collections and their hierarchical relationships.
// All sensitive collection metadata is encrypted client-side before being uploaded.
type CollectionDTO struct {
	// Identifiers
	// ID is the unique identifier for the collection in the cloud backend.
	ID gocql.UUID `bson:"_id" json:"id"`
	// OwnerID is the ID of the user who originally created and owns this collection.
	// The owner has administrative privileges by default.
	OwnerID gocql.UUID `bson:"owner_id" json:"owner_id"`

	// Encryption and Content Details
	// EncryptedName is the name of the collection, encrypted using the collection's unique key.
	// Stored and transferred in encrypted form.
	EncryptedName string `bson:"encrypted_name" json:"encrypted_name"`
	// CollectionType indicates the nature of the collection, either "folder" or "album".
	// Defined by CollectionTypeFolder and CollectionTypeAlbum constants.
	CollectionType string `bson:"collection_type" json:"collection_type"` // "folder" or "album"
	// EncryptedCollectionKey is the unique symmetric key used to encrypt the collection's data (like name and file metadata).
	// This key is encrypted with the owner's master key for storage and transmission,
	// allowing the owner's device to decrypt it using their master key.
	EncryptedCollectionKey *keys.EncryptedCollectionKey `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Sharing
	// Collection members (users with access)
	Members []*CollectionMembershipDTO `bson:"members" json:"members"`

	// Hierarchical structure fields
	// ParentID is the ID of the parent collection if this is a subcollection.
	// It is omitted (nil) for root collections. Used to reconstruct the hierarchy.
	ParentID gocql.UUID `bson:"parent_id,omitempty" json:"parent_id,omitempty"` // Parent collection ID, not stored for root collections
	// AncestorIDs is an array containing the IDs of all parent collections up to the root.
	// This field is used for efficient querying and traversal of the collection hierarchy without joins.
	AncestorIDs []gocql.UUID `bson:"ancestor_ids,omitempty" json:"ancestor_ids,omitempty"` // Array of ancestor IDs for efficient querying
	// Children is an optional field for representing subcollections embedded directly within this DTO.
	// This is primarily used for tree-like structures in certain API responses or operations, not for persistent storage in this format.
	// Recursive embedding of the same type.
	Children []*CollectionDTO `bson:"children,omitempty" json:"children,omitempty"`

	// Ownership, timestamps and conflict resolution
	// CreatedAt is the timestamp when the collection was initially created.
	// Recorded on the local device and synced.
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	// CreatedByUserID is the ID of the user who created this collection.
	CreatedByUserID gocql.UUID `bson:"created_by_user_id" json:"created_by_user_id"`
	// ModifiedAt is the timestamp of the last modification to the collection's metadata or content.
	// Updated on the local device and synced.
	ModifiedAt       time.Time  `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID gocql.UUID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	// The current version of the collection.
	Version uint64 `bson:"version" json:"version"`

	// State management
	State            string    `bson:"state" json:"state"`                         // active, deleted, archived
	TombstoneVersion uint64    `bson:"tombstone_version" json:"tombstone_version"` // The `version` number that this collection was deleted at.
	TombstoneExpiry  time.Time `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// CollectionMembershipDTO represents a user's access to a collection in DTO format
type CollectionMembershipDTO struct {
	ID             gocql.UUID `bson:"_id" json:"id"`
	CollectionID   gocql.UUID `bson:"collection_id" json:"collection_id"`     // ID of the collection (redundant but helpful for queries)
	RecipientID    gocql.UUID `bson:"recipient_id" json:"recipient_id"`       // User receiving access
	RecipientEmail string     `bson:"recipient_email" json:"recipient_email"` // Email for display purposes
	GrantedByID    gocql.UUID `bson:"granted_by_id" json:"granted_by_id"`     // User who shared the collection

	EncryptedCollectionKey *keys.EncryptedCollectionKey `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Access details
	PermissionLevel string    `bson:"permission_level" json:"permission_level"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`

	// Sharing origin tracking
	IsInherited     bool       `bson:"is_inherited" json:"is_inherited"`                               // Tracks whether access was granted directly or inherited from a parent
	InheritedFromID gocql.UUID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"` // InheritedFromID identifies which parent collection granted this access
}
