// native/desktop/maplefile-cli/internal/domain/fileupload/models.go
package fileupload

import (
	"time"

	"github.com/gocql/gocql"
)

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	FileID       gocql.UUID
	CollectionID gocql.UUID
	OwnerID      gocql.UUID
}

// FileUploadResult represents the result of a file upload
type FileUploadResult struct {
	FileID             gocql.UUID
	UploadedAt         time.Time
	FileSizeBytes      int64
	ThumbnailSizeBytes int64
	Success            bool
	Error              error
}

// FileUploadProgress tracks the progress of an upload
type FileUploadProgress struct {
	FileID        gocql.UUID
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
