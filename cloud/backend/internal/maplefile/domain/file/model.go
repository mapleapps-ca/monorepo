// cloud/backend/internal/maplefile/domain/file/model.go
package file

import (
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// File represents an encrypted file stored in the system.
// The actual file content is stored in S3 while this entity
// maintains the minimal metadata needed.
type File struct {
	// Identifiers
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	OwnerID      primitive.ObjectID `bson:"owner_id" json:"owner_id"`

	// Encryption and Content Details
	// Encrypted metadata (JSON blob encrypted by client, contains file name, mime type, etc.)
	EncryptedMetadata string `bson:"encrypted_metadata" json:"encrypted_metadata"`
	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `bson:"encrypted_file_key" json:"encrypted_file_key"`
	// Version identifier for the encryption scheme used
	EncryptionVersion string `bson:"encryption_version" json:"encryption_version"`
	// Hash of the encrypted file for integrity checking
	EncryptedHash string `bson:"encrypted_hash" json:"encrypted_hash"`

	// File Storage Object Details
	// The path/key in S3 storage where the encrypted file is stored. Note: Not exposed to the client.
	EncryptedFileObjectKey string `bson:"encrypted_file_object_key" json:"-"`
	// Size of the encrypted file in bytes. Not an encrypted value. Used for accounting/billing.
	EncryptedFileSizeInBytes int64 `bson:"encrypted_file_size_in_bytes" json:"encrypted_file_size_in_bytes"`

	// Thumbnail Storage Object Details
	// The key in S3 storage where the encrypted thumbnail is stored (if exists). Note: Not exposed to the client.
	EncryptedThumbnailObjectKey string `bson:"encrypted_thumbnail_object_key" json:"-"`
	// Size of the encrypted thumbnail file in bytes. Not an encrypted value. Used for accounting/billing.
	EncryptedThumbnailSizeInBytes int64 `bson:"encrypted_thumbnail_size_in_bytes" json:"encrypted_thumbnail_size_in_bytes"`

	// Timestamps
	// When was this file uploaded
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	// When was this file last modified
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
}
