// cloud/backend/internal/maplefile/domain/file/interface.go
package file

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/sync"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileMetadataRepository defines the interface for interacting with file metadata storage.
// It handles operations related to storing, retrieving, updating, and deleting file information (metadata).
type FileMetadataRepository interface {
	// Create saves a single File metadata record to the storage.
	Create(file *File) error
	// CreateMany saves multiple File metadata records to the storage.
	CreateMany(files []*File) error
	// Get retrieves a single File metadata record by its unique identifier (ID).
	Get(id primitive.ObjectID) (*File, error)
	// GetWithAnyState retrieves a file regardless of its state (for admin operations)
	GetWithAnyState(id primitive.ObjectID) (*File, error)
	// GetByIDs retrieves multiple File metadata records by their unique identifiers (IDs).
	GetByIDs(ids []primitive.ObjectID) ([]*File, error)
	// GetByCollection retrieves all File metadata records associated with a specific collection ID.
	GetByCollection(collectionID primitive.ObjectID) ([]*File, error)
	// Update modifies an existing File metadata record in the storage.
	Update(file *File) error
	// Delete removes a single File metadata record by its unique identifier (ID) - now soft delete.
	Delete(id primitive.ObjectID) error
	// HardDelete permanently removes a file metadata record
	HardDelete(id primitive.ObjectID) error
	// DeleteMany removes multiple File metadata records by their unique identifiers (IDs) - now soft delete.
	DeleteMany(ids []primitive.ObjectID) error
	// HardDeleteMany permanently removes multiple file metadata records
	HardDeleteMany(ids []primitive.ObjectID) error
	// CheckIfExistsByID verifies if a File metadata record with the given ID exists in the storage.
	CheckIfExistsByID(id primitive.ObjectID) (bool, error)
	// CheckIfUserHasAccess determines if a specific user (userID) has access permissions for a given file (fileID).
	CheckIfUserHasAccess(fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error)
	GetByCreatedByUserID(createdByUserID primitive.ObjectID) ([]*File, error)
	GetByOwnerID(ownerID primitive.ObjectID) ([]*File, error)

	// State management operations
	Archive(id primitive.ObjectID) error
	Restore(id primitive.ObjectID) error

	// Sync Operations
	// GetFileSyncData retrieves file sync data with pagination for the specified user
	GetFileSyncData(ctx context.Context, userID primitive.ObjectID, cursor *sync.SyncCursor, limit int64) (*sync.FileSyncResponse, error)
}

// FileObjectStorageRepository defines the interface for interacting with the actual encrypted file data storage.
// It handles operations related to storing, retrieving, deleting, and generating access URLs for encrypted data.
type FileObjectStorageRepository interface {
	// StoreEncryptedData saves encrypted file data to the storage system. It takes the owner's ID,
	// the file's ID (metadata ID), and the encrypted byte slice. It returns the storage path
	// where the data was saved, or an error.
	StoreEncryptedData(ownerID string, fileID string, encryptedData []byte) (string, error)
	// GetEncryptedData retrieves encrypted file data from the storage system using its storage path.
	// It returns the encrypted data as a byte slice, or an error.
	GetEncryptedData(storagePath string) ([]byte, error)
	// DeleteEncryptedData removes encrypted file data from the storage system using its storage path.
	DeleteEncryptedData(storagePath string) error
	// GeneratePresignedDownloadURL creates a temporary, time-limited URL that allows direct download
	// of the file data located at the given storage path, with proper content disposition headers.
	GeneratePresignedDownloadURL(storagePath string, duration time.Duration) (string, error)
	// GeneratePresignedUploadURL creates a temporary, time-limited URL that allows clients to upload
	// encrypted file data directly to the storage system at the specified storage path.
	GeneratePresignedUploadURL(storagePath string, duration time.Duration) (string, error)
	// VerifyObjectExists checks if an object exists at the given storage path.
	VerifyObjectExists(storagePath string) (bool, error)
	// GetObjectSize returns the size in bytes of the object at the given storage path.
	GetObjectSize(storagePath string) (int64, error)
}
