// cloud/backend/internal/maplefile/domain/file/model.go
package file

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
)

// File represents an encrypted file entity stored in the backend database (MongoDB).
// This entity holds metadata and pointers to the actual file content and thumbnail,
// which are stored separately in S3. All sensitive file metadata and the file itself
// are encrypted client-side before being uploaded. The backend stores only encrypted
// data and necessary non-sensitive identifiers or sizes for management.
type File struct {
	// Identifiers
	// Unique identifier for this specific file entity.
	ID gocql.UUID `bson:"_id" json:"id"`
	// Identifier of the collection this file belongs to. Used for grouping and key management.
	CollectionID gocql.UUID `bson:"collection_id" json:"collection_id"`
	// Identifier of the user who owns this file.
	OwnerID gocql.UUID `bson:"owner_id" json:"owner_id"`

	// Encryption and Content Details
	// Client-side encrypted JSON blob containing file-specific metadata like the original file name,
	// MIME type, size of the *unencrypted* data, etc. Encrypted by the client using the file key.
	EncryptedMetadata string `bson:"encrypted_metadata" json:"encrypted_metadata"`
	// The file-specific data encryption key (DEK) used to encrypt the file content and metadata.
	// This key is encrypted by the client using the collection's key (a KEK). The backend
	// stores this encrypted key; only a user with access to the KEK can decrypt it.
	EncryptedFileKey keys.EncryptedFileKey `bson:"encrypted_file_key" json:"encrypted_file_key"`
	// Version identifier for the encryption scheme or client application version used to
	// encrypt this file. Useful for migration or compatibility checks.
	EncryptionVersion string `bson:"encryption_version" json:"encryption_version"`
	// Cryptographic hash of the *encrypted* file content stored in S3. Used for integrity
	// verification upon download *before* decryption.
	EncryptedHash string `bson:"encrypted_hash" json:"encrypted_hash"`

	// File Storage Object Details
	// The unique key or path within the S3 bucket where the main encrypted file content is stored.
	// This is an internal backend detail and is not exposed to the client API.
	EncryptedFileObjectKey string `bson:"encrypted_file_object_key" json:"-"`
	// The size of the *encrypted* file content stored in S3, in bytes. This size is not sensitive
	// and is used by the backend for storage accounting, billing, and transfer management.
	EncryptedFileSizeInBytes int64 `bson:"encrypted_file_size_in_bytes" json:"encrypted_file_size_in_bytes"`

	// Thumbnail Storage Object Details (Optional)
	// The unique key or path within the S3 bucket where the encrypted thumbnail image (if generated
	// and uploaded) is stored. Internal backend detail, not exposed to the client API.
	EncryptedThumbnailObjectKey string `bson:"encrypted_thumbnail_object_key" json:"-"`
	// The size of the *encrypted* thumbnail image stored in S3, in bytes. Used for accounting.
	// Value will be 0 if no thumbnail exists.
	EncryptedThumbnailSizeInBytes int64 `bson:"encrypted_thumbnail_size_in_bytes" json:"encrypted_thumbnail_size_in_bytes"`

	// Timestamps and conflict resolution
	// Timestamp when this file entity was created/uploaded.
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	// CreatedByUserID is the ID of the user who created this file.
	CreatedByUserID gocql.UUID `bson:"created_by_user_id" json:"created_by_user_id"`
	// Timestamp when this file entity's metadata or content was last modified.
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
	// ModifiedByUserID is the ID of the user whom has last modified this file.
	ModifiedByUserID gocql.UUID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	// The current version of the file.
	Version uint64 `bson:"version" json:"version"` // Every mutation (create, update, delete) is a versioned operation, keep track of the version number with this variable

	// State management.
	State            string    `bson:"state" json:"state"`                         // pending, active, deleted, archived
	TombstoneVersion uint64    `bson:"tombstone_version" json:"tombstone_version"` // The `version` number that this collection was deleted at.
	TombstoneExpiry  time.Time `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// FileSyncCursor represents cursor-based pagination for sync operations
type FileSyncCursor struct {
	LastModified time.Time  `json:"last_modified" bson:"last_modified"`
	LastID       gocql.UUID `json:"last_id" bson:"last_id"`
}

// FileSyncItem represents minimal file data for sync operations
type FileSyncItem struct {
	ID               gocql.UUID `json:"id" bson:"_id"`
	CollectionID     gocql.UUID `json:"collection_id" bson:"collection_id"`
	Version          uint64     `json:"version" bson:"version"`
	ModifiedAt       time.Time  `json:"modified_at" bson:"modified_at"`
	State            string     `json:"state" bson:"state"`
	TombstoneVersion uint64     `bson:"tombstone_version" json:"tombstone_version"`
	TombstoneExpiry  time.Time  `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// FileSyncResponse represents the response for file sync data
type FileSyncResponse struct {
	Files      []FileSyncItem  `json:"files"`
	NextCursor *FileSyncCursor `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}
