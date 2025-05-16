// monorepo/cloud/backend/internal/maplefile/domain/file/model.go
package file

import (
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
)

// File represents an encrypted file stored in the system
// The actual file content is stored in S3 while this entity
// maintains the minimal metadata needed
type File struct {
	ID           string `bson:"id" json:"id"`
	CollectionID string `bson:"collection_id" json:"collection_id"`
	OwnerID      string `bson:"owner_id" json:"owner_id"`

	// Encrypted file identifier (client-generated)
	FileID string `bson:"file_id" json:"file_id"`

	// The path/key in S3 storage where the encrypted file is stored
	StoragePath string `bson:"storage_path" json:"storage_path"`

	// Size of the encrypted file in bytes
	EncryptedSize int64 `bson:"encrypted_size" json:"encrypted_size"`

	// The original file size before encryption, encrypted with file key
	EncryptedOriginalSize string `bson:"encrypted_original_size" json:"encrypted_original_size"`

	// Encrypted metadata (JSON blob encrypted by client)
	// Contains file name, mime type, etc.
	EncryptedMetadata string `bson:"encrypted_metadata" json:"encrypted_metadata"`

	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `bson:"encrypted_file_key" json:"encrypted_file_key"`

	// Version identifier for the encryption scheme used
	EncryptionVersion string `bson:"encryption_version" json:"encryption_version"`

	// Hash of the encrypted file for integrity checking
	EncryptedHash string `bson:"encrypted_hash" json:"encrypted_hash"`

	// Thumbnail data encrypted with file key (if available)
	EncryptedThumbnail string `bson:"encrypted_thumbnail" json:"encrypted_thumbnail,omitempty"`

	// When was this file uploaded
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// When was this file last modified
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
}
