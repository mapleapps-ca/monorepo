// cloud/backend/internal/maplefile/repo/filemetadata/get.go
package filemetadata

import (
	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// Get file by ID
func (impl fileMetadataRepositoryImpl) Get(id gocql.UUID) (*dom_file.File, error) {
	return nil, nil
}

func (impl fileMetadataRepositoryImpl) GetWithAnyState(id gocql.UUID) (*dom_file.File, error) {
	return nil, nil
}

// GetByIDs gets files by their IDs
func (impl fileMetadataRepositoryImpl) GetByIDs(ids []gocql.UUID) ([]*dom_file.File, error) {
	return nil, nil
}

// GetByEncryptedFileID gets a file by its client-generated EncryptedFileID
func (impl fileMetadataRepositoryImpl) GetByEncryptedFileID(encryptedFileID string) (*dom_file.File, error) {
	return nil, nil
}

// GetByCollection gets all files in a collection
func (impl fileMetadataRepositoryImpl) GetByCollection(collectionID gocql.UUID) ([]*dom_file.File, error) {
	return nil, nil
}
