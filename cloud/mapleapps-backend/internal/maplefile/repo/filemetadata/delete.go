// cloud/backend/internal/maplefile/repo/filemetadata/delete.go
package filemetadata

import (
	"github.com/gocql/gocql"
)

func (impl fileMetadataRepositoryImpl) SoftDelete(id gocql.UUID) error {

	return nil
}

func (impl fileMetadataRepositoryImpl) SoftDeleteMany(ids []gocql.UUID) error {
	return nil
}

// Add hard delete method for permanent removal
func (impl fileMetadataRepositoryImpl) HardDelete(id gocql.UUID) error {

	return nil
}

func (impl fileMetadataRepositoryImpl) HardDeleteMany(ids []gocql.UUID) error {

	return nil
}
