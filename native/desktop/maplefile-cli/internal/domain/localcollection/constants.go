// monorepo/native/desktop/maplefile-cli/internal/domain/localcollection/constants.go
package collection

// Constants for collection types
const (
	CollectionTypeFolder = "folder"
	CollectionTypeAlbum  = "album"
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
)
