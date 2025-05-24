// monorepo/native/desktop/maplefile-cli/internal/repo/file/utils.go
package file

import "fmt"

// generateKey creates a storage key for a file
func (r *fileRepository) generateKey(id string) string {
	return fmt.Sprintf("%s%s", fileKeyPrefix, id)
}
