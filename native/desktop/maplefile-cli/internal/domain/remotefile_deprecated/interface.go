// internal/domain/remotefile/interface.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RemoteFileRepository defines the interface for interacting with files on the cloud cloud backend
type RemoteFileRepository interface {
	Create(ctx context.Context, request *RemoteCreateFileRequest) (*RemoteFileResponse, error)
	Fetch(ctx context.Context, id primitive.ObjectID) (*RemoteFile, error)
	GetByCloudID(ctx context.Context, cloudID primitive.ObjectID) (*RemoteFile, error)
	ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*RemoteFile, error)
	List(ctx context.Context, filter RemoteFileFilter) ([]*RemoteFile, error)
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Upload/download operations
	GetDownloadURL(ctx context.Context, fileID primitive.ObjectID) (string, error)
	UploadFileByLocalID(ctx context.Context, localID primitive.ObjectID, data []byte) error
	UploadFileByCloudID(ctx context.Context, cloudID primitive.ObjectID, data []byte) error
	DownloadFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error)
}

// RemoteFileFilter defines filtering options for listing cloud files
type RemoteFileFilter struct {
	CollectionID *primitive.ObjectID `json:"collection_id,omitempty"`
	OwnerID      *primitive.ObjectID `json:"owner_id,omitempty"`
	Status       *FileStatus         `json:"status,omitempty"`
}
