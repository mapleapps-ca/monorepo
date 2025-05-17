// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/utils.go
package collection

import "fmt"

// generateKey creates a storage key for a collection
func (r *localcollectionRepository) generateKey(id string) string {
	return fmt.Sprintf("%s%s", collectionKeyPrefix, id)
}
