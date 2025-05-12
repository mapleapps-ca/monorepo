// cloud/backend/internal/papercloud/repo/file/get.go
package file

import (
	"fmt"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
)

// Get implements the FileRepository.Get method
func (repo *fileRepositoryImpl) Get(id string) (*dom_file.File, error) {
	return repo.metadata.Get(id)
}

// GetByCollection implements the FileRepository.GetByCollection method
func (repo *fileRepositoryImpl) GetByCollection(collectionID string) ([]*dom_file.File, error) {
	return repo.metadata.GetByCollection(collectionID)
}

// GetEncryptedData implements the FileRepository.GetEncryptedData method
func (repo *fileRepositoryImpl) GetEncryptedData(fileID string) ([]byte, error) {
	// Get file metadata to ensure it exists and get storage path
	file, err := repo.metadata.Get(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	// Retrieve the encrypted data from S3
	return repo.storage.GetEncryptedData(file.StoragePath)
}
