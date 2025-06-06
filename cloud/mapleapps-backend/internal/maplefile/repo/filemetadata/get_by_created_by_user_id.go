// cloud/backend/internal/maplefile/repo/filemetadata/get_by_created_by_user_id.go
package filemetadata

import (
	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// GetByCreatedByUserID gets all files created by a specific user
func (impl fileMetadataRepositoryImpl) GetByCreatedByUserID(createdByUserID gocql.UUID) ([]*dom_file.File, error) {
	return nil, nil
}
