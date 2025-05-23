// native/desktop/maplefile-cli/internal/domain/localfile/model.go
package localfile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// LocalFile represents a file on the user's local device.
type LocalFile struct {
	// Identifiers
	// Local primary key
	ID primitive.ObjectID `json:"id"`
	// Remote reference ID - stores the unique identifier of the corresponding remote file. This gets updated when the file is synced with the remote server.
	RemoteID primitive.ObjectID `json:"remote_id,omitempty"`
	// Collection this file belongs to
	CollectionID primitive.ObjectID `json:"collection_id"`
	// Owner of the file
	OwnerID primitive.ObjectID `json:"owner_id"`

	// Encryption, Decryption and Content Details
	// Client-side encrypted JSON blob containing file-specific metadata like the original file name,
	// MIME type, size of the *unencrypted* data, etc. Encrypted by the client using the file key.
	EncryptedMetadata string `json:"encrypted_metadata"`
	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `json:"encrypted_file_key"`
	// Version identifier for the encryption scheme used
	EncryptionVersion string `json:"encryption_version"`
	// Hash of the encrypted file for integrity checking
	EncryptedHash string `json:"encrypted_hash"`
	// Decrypted metadata for local use
	DecryptedName     string `json:"decrypted_name"`
	DecryptedMimeType string `json:"decrypted_mime_type"`

	// Encrypted File Storage Details
	// The path on the local filesystem where the encrypted file is stored
	EncryptedFilePath string `json:"encrypted_file_path"`
	// Size of the encrypted file in bytes. To be used for accounting and billing purposes.
	EncryptedFileSize int64 `json:"encrypted_file_size"`

	// Decrypted File Storage Details
	// The path on the local filesystem where the decrypted file is stored
	DecryptedFilePath string `json:"decrypted_file_path"`
	// Size of the decrypted file in bytes.
	DecryptedFileSize int64 `json:"decrypted_file_size"`

	// Encrypted Thumbnail Storage Details
	// The path where the thumbnail is stored locally (if it exists)
	EncryptedThumbnailPath string `json:"encrypted_thumbnail_path,omitempty"`
	// Size of the encrypted thumbnail in bytes. To be used for accounting and billing purposes.
	EncryptedThumbnailSize int64 `json:"encrypted_thumbnai_size"`

	// Decrypted Thumbnail Storage Details
	// The path where the thumbnail is stored locally (if it exists)
	DecryptedThumbnailPath string `json:"decrypted_thumbnail_path,omitempty"`
	// Size of the decrypted thumbnail in bytes. To be used for accounting and billing purposes.
	DecryptedThumbnailSize int64 `json:"decrypted_thumbnail_size"`

	// Fields for tracking synchronization state
	LastSyncedAt      time.Time  `json:"last_synced_at"`
	IsModifiedLocally bool       `json:"is_modified_locally"`
	SyncStatus        SyncStatus `json:"sync_status"`
	// Controls which file versions are kept (encrypted, decrypted, or both)
	StorageMode string `json:"storage_mode"`

	// Timestamps and conflict resolution
	// When was this file created locally
	CreatedAt time.Time `json:"created_at"`
	// CreatedByUserID is the ID of the user who created this file.
	CreatedByUserID primitive.ObjectID `json:"created_by_user_id"`
	// When was this file last modified
	ModifiedAt time.Time `json:"modified_at"`
	// ModifiedByUserID is the ID of the user whom has last modified this file.
	ModifiedByUserID primitive.ObjectID `json:"modified_by_user_id"`
	// The current version of the file.
	Version uint64 `json:"version"`
}
