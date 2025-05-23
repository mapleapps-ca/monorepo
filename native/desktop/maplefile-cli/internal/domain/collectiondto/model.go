// monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto/model.go
package collectiondto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

const (
	// CollectionTypeFolder represents a collection used to store files and other collections (folders or albums).
	CollectionTypeFolder = "folder"
	// CollectionTypeAlbum represents a collection primarily intended for storing media files,
	// often with specific metadata or display characteristics relevant to albums.
	CollectionTypeAlbum = "album"

	// Permission levels define the access rights users have to a collection.
	// These levels dictate what actions a user can perform within the collection (e.g., viewing, adding, deleting files/subcollections, managing members).

	// CollectionPermissionReadOnly grants users the ability to view the contents of the collection (files and subcollections)
	// and their metadata, but not modify them or the collection itself.
	CollectionPermissionReadOnly = "read_only"
	// CollectionPermissionReadWrite grants users the ability to view, add, modify, and delete
	// files and subcollections within the collection. They cannot manage collection members or delete the collection itself.
	CollectionPermissionReadWrite = "read_write"
	// CollectionPermissionAdmin grants users full control over the collection, including
	// all read/write operations, managing collection members (sharing/unsharing), and deleting the collection.
	CollectionPermissionAdmin = "admin"
)

// CollectionDTO represents a Data Transfer Object (DTO)
// used for transferring collection (folder or album) data between the local device and the cloud server.
// This data is end-to-end encrypted (E2EE) on the local device before transmission.
// The cloud server stores this encrypted data but cannot decrypt it.
// On the local device, this data is decrypted for use and storage (not stored in this encrypted DTO format locally).
// It can represent both root collections and embedded subcollections.
type CollectionDTO struct {
	// Identifiers
	// ID is the unique identifier for the collection in the cloud backend..
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// OwnerID is the ID of the user who originally created and owns this collection.
	// The owner has administrative privileges by default.
	OwnerID primitive.ObjectID `bson:"owner_id" json:"owner_id"`

	// Encryption and Content Details
	// EncryptedName is the name of the collection, encrypted using the collection's unique key.
	// Stored and transferred in encrypted form.
	EncryptedName string `bson:"encrypted_name" json:"encrypted_name"`
	// Type indicates the nature of the collection, either "folder" or "album".
	// Defined by CollectionTypeFolder and CollectionTypeAlbum constants.
	Type string `bson:"type" json:"type"` // "folder" or "album"
	// EncryptedCollectionKey is the unique symmetric key used to encrypt the collection's data (like name and file metadata).
	// This key is encrypted with the owner's master key for storage and transmission,
	// allowing the owner's device to decrypt it using their master key.
	EncryptedCollectionKey *keys.EncryptedCollectionKey `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Sharing
	// Members is a list of CollectionMembershipDTOs representing users who have access to this collection.
	// This list includes the owner and any users with whom the collection has been shared.
	Members []*CollectionMembershipDTO `bson:"members" json:"members"`

	// Hierarchical structure fields
	// ParentID is the ID of the parent collection if this is a subcollection.
	// It is omitted (nil) for root collections. Used to reconstruct the hierarchy.
	ParentID primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"` // Parent collection ID, not stored for root collections
	// AncestorIDs is an array containing the IDs of all parent collections up to the root.
	// This field is used for efficient querying and traversal of the collection hierarchy without joins.
	AncestorIDs []primitive.ObjectID `bson:"ancestor_ids,omitempty" json:"ancestor_ids,omitempty"` // Array of ancestor IDs for efficient querying
	// EncryptedPathSegments is an array of the encrypted names of the collection and its ancestors, from root to this collection.
	// This allows querying or sorting based on the path without decrypting the entire hierarchy.
	EncryptedPathSegments []string `bson:"encrypted_path_segments" json:"encrypted_path_segments"`
	// Children is an optional field for representing subcollections embedded directly within this DTO.
	// This is primarily used for tree-like structures in certain API responses or operations, not for persistent storage in this format.
	// Recursive embedding of the same type.
	Children []*CollectionDTO `bson:"children,omitempty" json:"children,omitempty"`

	// Ownership, timestamps and conflict resolution
	// CreatedAt is the timestamp when the collection was initially created.
	// Recorded on the local device and synced.
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	// CreatedByUserID is the ID of the user who created this file.
	CreatedByUserID primitive.ObjectID `json:"created_by_user_id"`
	// ModifiedAt is the timestamp of the last modification to the collection's metadata or content.
	// Updated on the local device and synced.
	ModifiedAt       time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID primitive.ObjectID `json:"modified_by_user_id"`
	// The current version of the file.
	Version uint64 `json:"version"`
}

// CollectionMembershipDTO represents a user's access to a collection.
// Each instance indicates that a specific RecipientID has a certain PermissionLevel for a given CollectionID.
type CollectionMembershipDTO struct {
	// ID is the unique identifier for this specific membership record in the cloud database.
	ID primitive.ObjectID `bson:"_id" json:"id"`

	// CollectionID is the ID of the collection to which this membership applies.
	// Redundant with the parent CollectionDTO's ID but useful for direct queries on membership records.
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"` // ID of the collection (redundant but helpful for queries)

	// RecipientID is the ID of the user who is granted access through this membership.
	RecipientID primitive.ObjectID `bson:"recipient_id" json:"recipient_id"` // User receiving access

	// RecipientEmail is the email address associated with the RecipientID.
	// Included for display purposes (e.g., showing who a collection is shared with) and stored unencrypted.
	RecipientEmail string `bson:"recipient_email" json:"recipient_email"` // Email for display purposes

	// GrantedByID is the ID of the user who created this membership record (i.e., performed the sharing action).
	// Useful for auditing and tracking the origin of sharing.
	GrantedByID primitive.ObjectID `bson:"granted_by_id" json:"granted_by_id"` // User who shared the collection

	// EncryptedCollectionKey is the collection's symmetric key, specifically encrypted for the RecipientID.
	// This key is encrypted using the recipient's public key (box_seal), allowing only the recipient to decrypt it using their private key.
	// This ensures that only authorized recipients can access the collection's data.
	EncryptedCollectionKey []byte `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Access details

	// PermissionLevel indicates the level of access the RecipientID has to the CollectionID.
	// Defined by CollectionPermissionReadOnly, CollectionPermissionReadWrite, CollectionPermissionAdmin constants.
	PermissionLevel string `bson:"permission_level" json:"permission_level"`

	// CreatedAt is the timestamp when this specific membership was created (i.e., when access was granted).
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// Sharing origin tracking

	// IsInherited is a flag indicating whether this access was granted directly to this collection (false)
	// or inherited from a parent collection due to sharing at a higher level in the hierarchy (true).
	IsInherited bool `bson:"is_inherited" json:"is_inherited"` // Tracks whether access was granted directly or inherited from a parent

	// InheritedFromID is the ID of the parent collection from which this membership was inherited, if IsInherited is true.
	// It is omitted (nil) if the access was granted directly to this collection.
	InheritedFromID primitive.ObjectID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"` // InheritedFromID identifies which parent collection granted this access
}
