// internal/domain/remotefile/constants.go
package remotefile

// FileStatus defines the status of a file in the remote system
type FileStatus int

const (
	// FileStatusPending indicates the file is being uploaded or processed
	FileStatusPending FileStatus = iota

	// FileStatusAvailable indicates the file is available for download
	FileStatusAvailable

	// FileStatusError indicates there was an error processing the file
	FileStatusError
)
