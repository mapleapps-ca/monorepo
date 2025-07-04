package file

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// File represents a file on the user's local device.
type File struct {
	// Identifiers
	// ID is the unique identifier of the corresponding cloud file set by the cloud server. This gets updated when the file is synced with the cloud server.
	ID gocql.UUID `json:"id" bson:"id"`
	// Collection this file belongs to
	CollectionID gocql.UUID `json:"collection_id" bson:"collection_id"`
	// Owner of the file
	OwnerID gocql.UUID `json:"owner_id" bson:"owner_id"`

	// Encryption, Decryption and Content Details
	// Client-side encrypted JSON blob containing file-specific metadata like the original file name,
	// MIME type, size of the *unencrypted* data, etc. Encrypted by the client using the file key.
	EncryptedMetadata string `json:"encrypted_metadata" bson:"encrypted_metadata"`
	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `json:"encrypted_file_key" bson:"encrypted_file_key"`
	// Version identifier for the encryption scheme used
	EncryptionVersion string `json:"encryption_version" bson:"encryption_version"`
	// Hash of the encrypted file for integrity checking
	EncryptedHash string `json:"encrypted_hash" bson:"encrypted_hash"`
	// Decrypted metadata for local use (client device side only)
	Name     string        `json:"name" bson:"name"`
	MimeType string        `json:"mime_type" bson:"mime_type"`
	Metadata *FileMetadata `json:"metadata" bson:"metadata"`

	// Encrypted File Storage Details
	// The path on the local filesystem where the encrypted file is stored
	EncryptedFilePath string `json:"encrypted_file_path" bson:"encrypted_file_path"`
	// Size of the encrypted file in bytes. To be used for accounting and billing purposes.
	EncryptedFileSize int64 `json:"encrypted_file_size" bson:"encrypted_file_size"`

	// Decrypted File Storage Details (client device side only)
	// The path on the local filesystem where the decrypted file is stored
	FilePath string `json:"file_path" bson:"file_path"`
	// Size of the decrypted file in bytes.
	FileSize int64 `json:"file_size" bson:"file_size"`

	// Encrypted Thumbnail Storage Details
	// The path where the thumbnail is stored locally (if it exists)
	EncryptedThumbnailPath string `json:"encrypted_thumbnail_path,omitempty" bson:"encrypted_thumbnail_path,omitempty"`
	// Size of the encrypted thumbnail in bytes. To be used for accounting and billing purposes.
	EncryptedThumbnailSize int64 `json:"encrypted_thumbnai_size" bson:"encrypted_thumbnai_size"`

	// Decrypted Thumbnail Storage Details (client device side only)
	// The path where the thumbnail is stored locally (if it exists)
	ThumbnailPath string `json:"thumbnail_path,omitempty" bson:"thumbnail_path,omitempty"`
	// Size of the decrypted thumbnail in bytes. To be used for accounting and billing purposes.
	ThumbnailSize int64 `json:"thumbnail_size" bson:"thumbnail_size"`

	// Fields for tracking synchronization state
	LastSyncedAt time.Time  `json:"last_synced_at" bson:"last_synced_at"`
	SyncStatus   SyncStatus `json:"sync_status" bson:"sync_status"`
	// Controls which file versions are kept (encrypted, decrypted, or both) (client device side only)
	StorageMode string `json:"storage_mode" bson:"storage_mode"`

	// Timestamps and conflict resolution
	// When was this file created locally
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	// CreatedByUserID is the ID of the user who created this file.
	CreatedByUserID gocql.UUID `json:"created_by_user_id" bson:"created_by_user_id"`
	// When was this file last modified
	ModifiedAt time.Time `json:"modified_at" bson:"modified_at"`
	// ModifiedByUserID is the ID of the user whom has last modified this file.
	ModifiedByUserID gocql.UUID `json:"modified_by_user_id" bson:"modified_by_user_id"`
	// The current version of the file.
	Version uint64 `bson:"version" json:"version"` // Every mutation (create, update, delete, etc) is a versioned operation, keep track of the version number with this variable

	// State management.
	State            string    `bson:"state" json:"state"`                         // pending, active, deleted, archived
	TombstoneVersion uint64    `bson:"tombstone_version" json:"tombstone_version"` // The `version` number that this collection was deleted at.
	TombstoneExpiry  time.Time `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// FileMetadata represents decrypted file metadata
type FileMetadata struct {
	Name                   string `json:"name"`
	MimeType               string `json:"mime_type"`
	Size                   int64  `json:"size"`
	Created                int64  `json:"created"`
	FileExtension          string `json:"file_extension"`
	EncryptedFilePath      string `json:"encrypted_file_path" bson:"encrypted_file_path"`
	EncryptedFileSize      int64  `json:"encrypted_file_size" bson:"encrypted_file_size"`
	DecryptedFilePath      string `json:"decrypted_file_path" bson:"decrypted_file_path"`
	DecryptedFileSize      int64  `json:"decrypted_file_size" bson:"decrypted_file_size"`
	EncryptedThumbnailPath string `json:"encrypted_thumbnail_path,omitempty" bson:"encrypted_thumbnail_path,omitempty"`
	EncryptedThumbnailSize int64  `json:"encrypted_thumbnai_size" bson:"encrypted_thumbnai_size"`
	DecryptedThumbnailPath string `json:"decrypted_thumbnail_path,omitempty" bson:"decrypted_thumbnail_path,omitempty"`
	DecryptedThumbnailSize int64  `json:"decrypted_thumbnail_size" bson:"decrypted_thumbnail_size"`
}
