package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileRepository defines the interface for interacting with file database
type FileRepository interface {
	// Create saves a single File record to the storage.
	Create(ctx context.Context, file *File) error
	// CreateMany saves multiple File records to the storage.
	CreateMany(ctx context.Context, files []*File) error
	// Get retrieves a single File record by its unique identifier (ID).
	Get(ctx context.Context, id primitive.ObjectID) (*File, error)
	// GetByIDs retrieves multiple File records by their unique identifiers (IDs).
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*File, error)
	// GetByCollection retrieves all File records associated with a specific collection ID.
	GetByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*File, error)
	// Update modifies an existing File record in the storage.
	Update(ctx context.Context, file *File) error
	// Delete removes a single File record by its unique identifier (ID).
	Delete(ctx context.Context, id primitive.ObjectID) error
	// DeleteMany removes multiple File records by their unique identifiers (IDs).
	DeleteMany(ctx context.Context, ids []primitive.ObjectID) error
	// CheckIfExistsByID verifies if a File record with the given ID exists in the storage.
	CheckIfExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error)
	// CheckIfUserHasAccess determines if a specific user (userID) has access permissions for a given file (fileID).
	CheckIfUserHasAccess(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error)

	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// FileFilter defines filtering options for listing local files
type FileFilter struct {
	CollectionID *primitive.ObjectID `json:"collection_id,omitempty"`
	SyncStatus   *SyncStatus         `json:"sync_status,omitempty"`
	NameContains *string             `json:"name_contains,omitempty"`
	MimeType     *string             `json:"mime_type,omitempty"`
}
