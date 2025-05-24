// monorepo/native/desktop/maplefile-cli/internal/domain/collection/constants.go
package collection

// Constants for collection types
const (
	// CollectionTypeFolder represents a collection used to store files and other collections (folders or albums).
	CollectionTypeFolder = "folder"
	// CollectionTypeAlbum represents a collection primarily intended for storing media files,
	// often with specific metadata or display characteristics relevant to albums.
	CollectionTypeAlbum = "album"
)

// SyncStatus defines the synchronization status of a collection
type SyncStatus int

const (
	// SyncStatusLocalOnly indicates the collection exists only locally
	SyncStatusLocalOnly SyncStatus = iota

	// SyncStatusCloudOnly indicates the collection exists only in the cloud
	SyncStatusCloudOnly

	// SyncStatusSynced indicates the collection exists both locally and in the cloud and is in sync
	SyncStatusSynced

	// SyncStatusModifiedLocally indicates the collection exists in both places but has local changes
	SyncStatusModifiedLocally

	// Permission levels define the access rights users have to a collection.
	// These levels dictate what actions a user can perform within the collection (e.g., viewing, adding, deleting files/subcollections, managing members).

	// CollectionPermissionReadOnly grants users the ability to view the contents of the collection (files and subcollections)
	// and their metadata, but not modify them or the collection itself.
	CollectionPermissionReadOnly = "read_only"
	// CollectionPermissionReadWrite grants users the ability to view, add, modify, and delete
	// files and subcollections within the collection. They cannot manage collection members or delete the collection itself.
	CollectionPermissionReadWrite = "read_write"
	// CollectionPermissionAdmin grants users full control over the collection, including
	// all read/write operations, managing collection members (sharing/unsharing), and deleting the collection.
	CollectionPermissionAdmin = "admin"
)
