// cloud/backend/internal/maplefile/domain/file/constants.go
package file

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
