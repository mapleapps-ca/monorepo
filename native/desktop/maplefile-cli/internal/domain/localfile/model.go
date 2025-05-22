// internal/domain/localfile/model.go
package localfile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// LocalFile represents a file on the user's local device.
type LocalFile struct {
	// Local primary key
	ID primitive.ObjectID `json:"id"`

	// Remote reference ID - stores the ID of the corresponding remote file
	RemoteID primitive.ObjectID `json:"remote_id,omitempty"`

	// Collection this file belongs to
	CollectionID primitive.ObjectID `json:"collection_id"`

	// Owner of the file
	OwnerID primitive.ObjectID `json:"owner_id"`

	// Encrypted file identifier (client-generated)
	EncryptedFileID string `json:"encrypted_file_id"`

	// Encrypted metadata (JSON blob encrypted by client)
	// Contains file name, mime type, etc.
	EncryptedMetadata string `json:"encrypted_metadata"`

	// Decrypted metadata for local use
	DecryptedName     string `json:"decrypted_name"`
	DecryptedMimeType string `json:"decrypted_mime_type"`

	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `json:"encrypted_file_key"`

	// Version identifier for the encryption scheme used
	EncryptionVersion string `json:"encryption_version"`

	// Hash of the encrypted file for integrity checking
	EncryptedHash string `json:"encrypted_hash"`

	// The path where the thumbnail is stored locally (if it exists)
	LocalThumbnailPath string `json:"local_thumbnail_path,omitempty"`

	// When was this file created locally
	CreatedAt time.Time `json:"created_at"`

	// When was this file last modified
	ModifiedAt time.Time `json:"modified_at"`

	// Fields for tracking synchronization state
	LastSyncedAt      time.Time  `json:"last_synced_at"`
	IsModifiedLocally bool       `json:"is_modified_locally"`
	SyncStatus        SyncStatus `json:"sync_status"`

	// The path on the local filesystem where the encrypted file is stored
	EncryptedFilePath string `json:"encrypted_file_path"`

	// Size of the encrypted file in bytes. To be used for accounting and billing purposes.
	EncryptedFileSize int64 `json:"encrypted_file_size"`

	// The path on the local filesystem where the decrypted file is stored
	DecryptedFilePath string `json:"decrypted_file_path"`

	// Size of the decrypted file in bytes.
	DecryptedFileSize int64 `json:"decrypted_file_size"`

	// Controls which file versions are kept (encrypted, decrypted, or both)
	StorageMode string `json:"storage_mode"`
}
