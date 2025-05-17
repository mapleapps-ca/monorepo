// internal/domain/localfile/interface.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LocalFileRepository defines the interface for interacting with local files
type LocalFileRepository interface {
	Create(ctx context.Context, file *LocalFile) error
	Save(ctx context.Context, file *LocalFile) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*LocalFile, error)
	GetByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*LocalFile, error)
	GetByEncryptedFileID(ctx context.Context, encryptedFileID string) (*LocalFile, error)
	List(ctx context.Context, filter LocalFileFilter) ([]*LocalFile, error)
	ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*LocalFile, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// LocalFileFilter defines filtering options for listing local files
type LocalFileFilter struct {
	CollectionID *primitive.ObjectID `json:"collection_id,omitempty"`
	RemoteID     *primitive.ObjectID `json:"remote_id,omitempty"`
	SyncStatus   *SyncStatus         `json:"sync_status,omitempty"`
	NameContains *string             `json:"name_contains,omitempty"`
	MimeType     *string             `json:"mime_type,omitempty"`
}
