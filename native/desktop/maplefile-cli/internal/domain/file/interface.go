package file

import (
	"context"

	"github.com/gocql/gocql"
)

// FileRepository defines the interface for interacting with file database
type FileRepository interface {
	// Create saves a single File record to the storage.
	Create(ctx context.Context, file *File) error
	// CreateMany saves multiple File records to the storage.
	CreateMany(ctx context.Context, files []*File) error
	// Get retrieves a single File record by its unique identifier (ID).
	Get(ctx context.Context, id gocql.UUID) (*File, error)
	// GetByIDs retrieves multiple File records by their unique identifiers (IDs).
	GetByIDs(ctx context.Context, ids []gocql.UUID) ([]*File, error)
	// GetByCollection retrieves all File records associated with a specific collection ID.
	GetByCollection(ctx context.Context, collectionID gocql.UUID) ([]*File, error)
	// Update modifies an existing File record in the storage.
	Update(ctx context.Context, file *File) error
	// Delete removes a single File record by its unique identifier (ID).
	Delete(ctx context.Context, id gocql.UUID) error
	// DeleteMany removes multiple File records by their unique identifiers (IDs).
	DeleteMany(ctx context.Context, ids []gocql.UUID) error
	// CheckIfExistsByID verifies if a File record with the given ID exists in the storage.
	CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error)
	// CheckIfUserHasAccess determines if a specific user (userID) has access permissions for a given file (fileID).
	CheckIfUserHasAccess(ctx context.Context, fileID gocql.UUID, userID gocql.UUID) (bool, error)
	// SwapIDs will replace the oldID with the newID of a File record.
	SwapIDs(ctx context.Context, oldID gocql.UUID, newID gocql.UUID) error

	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// FileFilter defines filtering options for listing local files
type FileFilter struct {
	CollectionID *gocql.UUID `json:"collection_id,omitempty"`
	SyncStatus   *SyncStatus `json:"sync_status,omitempty"`
	NameContains *string     `json:"name_contains,omitempty"`
	MimeType     *string     `json:"mime_type,omitempty"`
}
