// internal/domain/localfile/constants.go
package localfile

// SyncStatus defines the synchronization status of a file
type SyncStatus int

const (
	// SyncStatusLocalOnly indicates the file exists only locally
	SyncStatusLocalOnly SyncStatus = iota

	// SyncStatusCloudOnly indicates the file exists only in the cloud
	SyncStatusCloudOnly

	// SyncStatusSynced indicates the file exists both locally and in the cloud and is in sync
	SyncStatusSynced

	// SyncStatusModifiedLocally indicates the file exists in both places but has local changes
	SyncStatusModifiedLocally
)

// LocalFileState constants define the encryption state of the file
const (
	LocalFileStateLocalAndDecrypted   = "local_and_decrypted"
	LocalFileStateLocalAndEncrypted   = "local_and_encrypted"
	LocalFileStateInCloudAndEncrypted = "in_cloud_and_encrypted"
)
