// monorepo/cloud/backend/internal/maplefile/domain/file/model.go
package file

import (
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// File represents an encrypted file stored in the system
// The actual file content is stored in S3 while this entity
// maintains the minimal metadata needed
type File struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	OwnerID      primitive.ObjectID `bson:"owner_id" json:"owner_id"`

	// Encrypted file identifier (client-generated)
	EncryptedFileID string `bson:"encrypted_file_id" json:"encrypted_file_id"`

	// The path/key in S3 storage where the encrypted file is stored
	FileObjectKey string `bson:"file_object_key" json:"file_object_key"`

	// Size of the file in bytes (with encryption overhead included). To be used for accounting and billing purposes.
	FileSize int64 `bson:"file_size" json:"file_size"`

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

	// The path/key in S3 storage where the encrypted file's thumbnail is stored (if it exist)
	ThumbnailObjectKey string `bson:"thumbnail_object_key" json:"thumbnail_object_key"`

	// When was this file uploaded
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// When was this file last modified
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
}
