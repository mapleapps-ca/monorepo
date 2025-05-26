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

const (
	// FileStatePending is the initial state of a file before it is uploaded.
	FileStatePending = "pending"
	// FileStateActive indicates that the file is fully uploaded and ready for use.
	FileStateActive = "active"
	// FileStateDeleted marks the file as deleted, but still accessible for a period but will eventually be permanently removed.
	FileStateDeleted = "deleted"
	// FileStateArchived indicates that the file is no longer accessible.
	FileStateArchived = "archived"
)
