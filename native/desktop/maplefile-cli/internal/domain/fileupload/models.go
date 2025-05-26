// native/desktop/maplefile-cli/internal/domain/fileupload/models.go
package fileupload

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	FileID       primitive.ObjectID
	CollectionID primitive.ObjectID
	OwnerID      primitive.ObjectID
}

// FileUploadResult represents the result of a file upload
type FileUploadResult struct {
	FileID             primitive.ObjectID
	UploadedAt         time.Time
	FileSizeBytes      int64
	ThumbnailSizeBytes int64
	Success            bool
	Error              error
}

// FileUploadProgress tracks the progress of an upload
type FileUploadProgress struct {
	FileID        primitive.ObjectID
	State         UploadState
	BytesUploaded int64
	TotalBytes    int64
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         error
}

// UploadState represents the state of an upload
type UploadState string

const (
	UploadStatePending    UploadState = "pending"
	UploadStateCreating   UploadState = "creating"
	UploadStateUploading  UploadState = "uploading"
	UploadStateCompleting UploadState = "completing"
	UploadStateCompleted  UploadState = "completed"
	UploadStateFailed     UploadState = "failed"
)
