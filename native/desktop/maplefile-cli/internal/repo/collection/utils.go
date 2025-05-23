// monorepo/native/desktop/maplefile-cli/internal/repo/collection/utils.go
package collection

import "fmt"

// generateKey creates a storage key for a collection
func (r *collectionRepository) generateKey(id string) string {
	return fmt.Sprintf("%s%s", collectionKeyPrefix, id)
}
