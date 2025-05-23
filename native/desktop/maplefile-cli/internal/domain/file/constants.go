// internal/domain/file/constants.go
package file

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

// Storage mode constants define which file versions to keep
const (
	StorageModeEncryptedOnly = "encrypted_only" // Only keep encrypted version (more secure)
	StorageModeDecryptedOnly = "decrypted_only" // Only keep decrypted version (not recommended)
	StorageModeHybrid        = "hybrid"         // Keep both versions (convenient)
)
