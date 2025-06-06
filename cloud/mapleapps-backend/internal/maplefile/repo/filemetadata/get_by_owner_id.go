// cloud/backend/internal/maplefile/repo/filemetadata/get_by_owner_id.go
package filemetadata

import (
	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// GetByOwnerID gets all files owned by a specific user
func (impl fileMetadataRepositoryImpl) GetByOwnerID(ownerID gocql.UUID) ([]*dom_file.File, error) {
	return nil, nil
}
