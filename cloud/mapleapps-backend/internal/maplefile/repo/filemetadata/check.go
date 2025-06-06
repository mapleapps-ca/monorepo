// cloud/backend/internal/maplefile/repo/filemetadata/check.go
package filemetadata

import (
	"github.com/gocql/gocql"
)

// CheckIfExistsByID checks if a file exists by ID
func (impl fileMetadataRepositoryImpl) CheckIfExistsByID(id gocql.UUID) (bool, error) {
	return false, nil
}

// CheckIfUserHasAccess checks if a user has access to a file
func (impl fileMetadataRepositoryImpl) CheckIfUserHasAccess(fileID gocql.UUID, userID gocql.UUID) (bool, error) {
	return false, nil
}
