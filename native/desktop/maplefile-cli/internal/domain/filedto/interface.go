// monorepo/native/desktop/maplefile-cli/internal/domain/filedto/interface.go
package filedto

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileDTORepository defines the interface for interacting with the cloud service
// to manage FileDTOs using the three-step upload process. These DTOs represent
// encrypted file data exchanged between the local device and the cloud server.
type FileDTORepository interface {
	// Three-Step File Upload Process

	// CreatePendingFileInCloud creates a pending file record in the cloud and returns
	// presigned URLs for uploading the file content directly to cloud storage.
	// This is Step 1 of the three-step upload process.
	CreatePendingFileInCloud(ctx context.Context, request *CreatePendingFileRequest) (*CreatePendingFileResponse, error)

	// UploadFileToCloud uploads the actual file content to cloud storage using
	// the presigned URL obtained from CreatePendingFile. This is Step 2 of the
	// three-step upload process. The actual upload is handled by this method.
	UploadFileToCloud(ctx context.Context, presignedURL string, fileData []byte) error

	// UploadThumbnailToCloud uploads the thumbnail content to cloud storage using
	// the presigned thumbnail URL. This is optional and part of Step 2.
	UploadThumbnailToCloud(ctx context.Context, presignedURL string, thumbnailData []byte) error

	// CompleteFileUploadInCloud completes the file upload process by notifying the cloud
	// service that the upload is finished. This transitions the file from 'pending'
	// to 'active' state. This is Step 3 of the three-step upload process.
	CompleteFileUploadInCloud(ctx context.Context, fileID primitive.ObjectID, request *CompleteFileUploadRequest) (*CompleteFileUploadResponse, error)

	// GetPresignedUploadURLFromCloud generates new presigned upload URLs for an existing file.
	// This can be used to re-upload or replace file content.
	GetPresignedUploadURLFromCloud(ctx context.Context, fileID primitive.ObjectID, request *GetPresignedUploadURLRequest) (*GetPresignedUploadURLResponse, error)

	// DownloadByIDFromCloud downloads a FileDTO by its unique identifier from the cloud service.
	DownloadByIDFromCloud(ctx context.Context, id primitive.ObjectID) (*FileDTO, error) // (Deprecated)

	// GetPresignedDownloadURLFromCloud generates presigned download URLs for an existing file.
	GetPresignedDownloadURLFromCloud(ctx context.Context, fileID primitive.ObjectID, request *GetPresignedDownloadURLRequest) (*GetPresignedDownloadURLResponse, error)

	// DownloadFileFromPresignedURL downloads file content from a presigned URL.
	DownloadFileFromPresignedURL(ctx context.Context, presignedURL string) ([]byte, error)

	// DownloadThumbnailFromPresignedURL downloads thumbnail content from a presigned URL.
	DownloadThumbnailFromPresignedURL(ctx context.Context, presignedURL string) ([]byte, error)

	// ListFromCloud lists FileDTOs from the cloud service based on the provided filter criteria.
	ListFromCloud(ctx context.Context, filter FileFilter) ([]*FileDTO, error)

	// DeleteByIDFromCloud deletes a FileDTO by its unique identifier from the cloud service.
	DeleteByIDFromCloud(ctx context.Context, id primitive.ObjectID) error
}

// Three-Step Upload Request/Response Types

// CreatePendingFileRequest represents the request to create a pending file record
type CreatePendingFileRequest struct {
	ID                           primitive.ObjectID `json:"id"`
	CollectionID                 primitive.ObjectID `json:"collection_id"`
	EncryptedMetadata            string             `json:"encrypted_metadata"`
	EncryptedFileKey             EncryptedFileKey   `json:"encrypted_file_key"`
	EncryptionVersion            string             `json:"encryption_version"`
	EncryptedHash                string             `json:"encrypted_hash"`
	ExpectedFileSizeInBytes      int64              `json:"expected_file_size_in_bytes"`
	ExpectedThumbnailSizeInBytes int64              `json:"expected_thumbnail_size_in_bytes,omitempty"`
}

// EncryptedFileKey represents an encrypted file key with nonce
type EncryptedFileKey struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
}

// CreatePendingFileResponse represents the response from creating a pending file
type CreatePendingFileResponse struct {
	File                    *FileDTO  `json:"file"`
	PresignedUploadURL      string    `json:"presigned_upload_url"`
	PresignedThumbnailURL   string    `json:"presigned_thumbnail_url,omitempty"`
	UploadURLExpirationTime time.Time `json:"upload_url_expiration_time"`
	Success                 bool      `json:"success"`
	Message                 string    `json:"message"`
}

// CompleteFileUploadRequest represents the request to complete a file upload
type CompleteFileUploadRequest struct {
	ActualFileSizeInBytes      int64 `json:"actual_file_size_in_bytes,omitempty"`
	ActualThumbnailSizeInBytes int64 `json:"actual_thumbnail_size_in_bytes,omitempty"`
	UploadConfirmed            bool  `json:"upload_confirmed,omitempty"`
	ThumbnailUploadConfirmed   bool  `json:"thumbnail_upload_confirmed,omitempty"`
}

// CompleteFileUploadResponse represents the response from completing a file upload
type CompleteFileUploadResponse struct {
	File                *FileDTO `json:"file"`
	Success             bool     `json:"success"`
	Message             string   `json:"message"`
	ActualFileSize      int64    `json:"actual_file_size"`
	ActualThumbnailSize int64    `json:"actual_thumbnail_size"`
	UploadVerified      bool     `json:"upload_verified"`
	ThumbnailVerified   bool     `json:"thumbnail_verified"`
}

// GetPresignedUploadURLRequest represents the request to get presigned upload URLs
type GetPresignedUploadURLRequest struct {
	URLDuration time.Duration `json:"url_duration,omitempty"` // Optional, defaults to 1 hour
}

// GetPresignedUploadURLResponse represents the response with presigned upload URLs
type GetPresignedUploadURLResponse struct {
	File                    *FileDTO  `json:"file"`
	PresignedUploadURL      string    `json:"presigned_upload_url"`
	PresignedThumbnailURL   string    `json:"presigned_thumbnail_url,omitempty"`
	UploadURLExpirationTime time.Time `json:"upload_url_expiration_time"`
	Success                 bool      `json:"success"`
	Message                 string    `json:"message"`
}

// FileFilter defines filtering options for listing FileDTOs.
type FileFilter struct {
	// CollectionID filters files that belong to the specified collection.
	CollectionID *primitive.ObjectID `json:"collection_id,omitempty"`
	// State filters files by their state (pending, active, deleted, archived).
	State string `json:"state,omitempty"`
}

// GetPresignedDownloadURLRequest represents the request to get presigned download URLs
type GetPresignedDownloadURLRequest struct {
	URLDuration time.Duration `json:"url_duration,omitempty"` // Optional, defaults to 1 hour
}

// GetPresignedDownloadURLResponse represents the response with presigned download URLs
type GetPresignedDownloadURLResponse struct {
	File                      *FileDTO  `json:"file"`
	PresignedDownloadURL      string    `json:"presigned_download_url"`
	PresignedThumbnailURL     string    `json:"presigned_thumbnail_url,omitempty"`
	DownloadURLExpirationTime time.Time `json:"download_url_expiration_time"`
	Success                   bool      `json:"success"`
	Message                   string    `json:"message"`
}

// Add
