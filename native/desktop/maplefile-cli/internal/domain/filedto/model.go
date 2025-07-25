// monorepo/native/desktop/maplefile-cli/internal/domain/filedto/model.go
package filedto

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// FileDTO represents a Data Transfer Object (DTO) used for transferring an encrypted file entity
// to and from a cloud service. This entity holds metadata and pointers to the
// actual file content and thumbnail, which are stored separately in S3. All sensitive file
// metadata and the file itself are encrypted client-side before being uploaded. The backend
// stores only encrypted data and necessary non-sensitive identifiers or sizes for management.
type FileDTO struct {
	// Identifiers
	// Unique identifier set by the cloud service.
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
	CreatedByUserID gocql.UUID `json:"created_by_user_id"`
	// Timestamp when this file entity's metadata or content was last modified.
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
	// ModifiedByUserID is the ID of the user whom has last modified this file.
	ModifiedByUserID gocql.UUID `json:"modified_by_user_id"`
	// The current version of the file.
	Version uint64 `json:"version"`

	// State management.
	State string `bson:"state" json:"state"` // pending, active, deleted, archived
}
