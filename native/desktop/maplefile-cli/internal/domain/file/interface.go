// native/desktop/maplefile-cli/internal/domain/file/interface.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileRepository defines the interface for interacting with local files
type FileRepository interface {
	Create(ctx context.Context, file *File) error

	Save(ctx context.Context, file *File) error

	GetByID(ctx context.Context, id primitive.ObjectID) (*File, error)
	GetByCloudID(ctx context.Context, cloudID primitive.ObjectID) (*File, error)

	List(ctx context.Context, filter FileFilter) ([]*File, error)
	ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*File, error)

	Delete(ctx context.Context, id primitive.ObjectID) error

	SaveEncryptedFileDataInternal(ctx context.Context, dataPath string, file *File, data []byte) error
	SaveDecryptedFileDataInternal(ctx context.Context, dataPath string, file *File, data []byte) error
	SaveHybridFileDataInternal(ctx context.Context, dataPath string, file *File, data []byte) error
	LoadDecryptedFileDataAtFilePath(ctx context.Context, decryptedFilePath string) ([]byte, error)
	LoadEncryptedFileDataAtFilePath(ctx context.Context, encryptedFilePath string) ([]byte, error)

	ImportFile(ctx context.Context, filePath string, file *File) error

	SaveThumbnail(ctx context.Context, file *File, thumbnailData []byte) error
	LoadThumbnail(ctx context.Context, file *File) ([]byte, error)

	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// FileFilter defines filtering options for listing local files
type FileFilter struct {
	CollectionID *primitive.ObjectID `json:"collection_id,omitempty"`
	CloudID      *primitive.ObjectID `json:"cloud_id,omitempty"`
	SyncStatus   *SyncStatus         `json:"sync_status,omitempty"`
	NameContains *string             `json:"name_contains,omitempty"`
	MimeType     *string             `json:"mime_type,omitempty"`
}
