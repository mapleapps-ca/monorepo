// native/desktop/maplefile-cli/internal/domain/remotefile/model.go
package remotefile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// RemoteFile represents a file stored in the cloud backend.
type RemoteFile struct {
	// Remote primary key
	ID primitive.ObjectID `json:"id"`

	// LocalID is the local file ID associated with this remote file.
	LocalID primitive.ObjectID `json:"local_id"`

	// Collection this file belongs to
	CollectionID primitive.ObjectID `json:"collection_id"`

	// Owner of the file
	OwnerID primitive.ObjectID `json:"owner_id"`

	// The path/key in S3 storage where the encrypted file is stored
	FileObjectKey string `json:"file_object_key"`

	// Size of the encrypted file in bytes
	EncryptedFileSize int64 `json:"encrypted_file_size"`

	// Encrypted metadata (JSON blob encrypted by client)
	// Contains file name, mime type, etc.
	EncryptedMetadata string `json:"encrypted_metadata"`

	// File-specific encryption key, encrypted with the collection key
	EncryptedFileKey keys.EncryptedFileKey `json:"encrypted_file_key"`

	// Version identifier for the encryption scheme used
	EncryptionVersion string `json:"encryption_version"`

	// Hash of the encrypted file for integrity checking
	EncryptedHash string `json:"encrypted_hash"`

	// The path/key in S3 storage where the encrypted file's thumbnail is stored (if it exists)
	ThumbnailObjectKey string `json:"thumbnail_object_key,omitempty"`

	// When was this file uploaded
	CreatedAt time.Time `json:"created_at"`

	// When was this file last modified
	ModifiedAt time.Time `json:"modified_at"`

	// Current status of the file in the remote system
	Status FileStatus `json:"status"`

	// Presigned download URL (if available)
	DownloadURL string `json:"download_url,omitempty"`

	// Download URL expiry time (if available)
	DownloadExpiry string `json:"download_expiry,omitempty"`
}

// RemoteCreateFileRequest represents the data needed to create a file
type RemoteCreateFileRequest struct {
	LocalFileID           primitive.ObjectID    `json:"local_file_id"`
	CollectionID          primitive.ObjectID    `json:"collection_id"`
	EncryptedFileSize     int64                 `json:"encrypted_file_size"`
	EncryptedOriginalSize string                `json:"encrypted_original_size"`
	EncryptedMetadata     string                `json:"encrypted_metadata"`
	EncryptedFileKey      keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion     string                `json:"encryption_version"`
	EncryptedHash         string                `json:"encrypted_hash,omitempty"`
}

// RemoteFileResponse represents the server's response when creating or fetching a file
type RemoteFileResponse struct {
	ID                    primitive.ObjectID    `json:"id"`
	CollectionID          primitive.ObjectID    `json:"collection_id"`
	OwnerID               primitive.ObjectID    `json:"owner_id"`
	FileObjectKey         string                `json:"file_object_key"`
	EncryptedFileSize     int64                 `json:"encrypted_file_size"`
	EncryptedOriginalSize string                `json:"encrypted_original_size"`
	EncryptedMetadata     string                `json:"encrypted_metadata"`
	EncryptedFileKey      keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion     string                `json:"encryption_version"`
	EncryptedHash         string                `json:"encrypted_hash"`
	ThumbnailObjectKey    string                `json:"thumbnail_object_key,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	ModifiedAt            time.Time             `json:"modified_at"`
	DownloadURL           string                `json:"download_url,omitempty"`    // Presigned URL for download
	DownloadExpiry        string                `json:"download_expiry,omitempty"` // Download URL expiry time
}
