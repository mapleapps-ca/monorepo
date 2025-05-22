// cloud/backend/internal/maplefile/repo/file/get.go
package file

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// Get implements the FileRepository.Get method
func (repo *fileRepositoryImpl) Get(id primitive.ObjectID) (*dom_file.File, error) {
	return repo.metadata.Get(id)
}

// GetMany implements the FileRepository.GetMany method
func (repo *fileRepositoryImpl) GetMany(ids []primitive.ObjectID) ([]*dom_file.File, error) {
	return repo.metadata.GetByIDs(ids)
}

// GetByCollection implements the FileRepository.GetByCollection method
func (repo *fileRepositoryImpl) GetByCollection(collectionID string) ([]*dom_file.File, error) {
	objectID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID format: %v", err)
	}
	return repo.metadata.GetByCollection(objectID)
}

// GetEncryptedData implements the FileRepository.GetEncryptedData method
func (repo *fileRepositoryImpl) GetEncryptedData(fileID primitive.ObjectID) ([]byte, error) {
	// Get file metadata to ensure it exists and get storage path
	file, err := repo.metadata.Get(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("encrypted file not found: %s", fileID.Hex())
	}

	// Check if file has been stored
	if file.EncryptedFileObjectKey == "" {
		return nil, fmt.Errorf("encrypted file data not yet stored: %s", fileID.Hex())
	}

	// Retrieve the encrypted data from S3
	return repo.storage.GetEncryptedData(file.EncryptedFileObjectKey)
}
