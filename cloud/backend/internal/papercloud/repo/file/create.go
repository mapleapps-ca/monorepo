// cloud/backend/internal/papercloud/repo/file/create.go
package file

import (
	"fmt"

	"github.com/google/uuid"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
)

// Create implements the FileRepository.Create method
func (repo *fileRepositoryImpl) Create(file *dom_file.File) error {
	// Generate ID if not provided
	if file.ID == "" {
		file.ID = uuid.New().String()
	}

	// If FileID is not set, generate one - ideally this comes from client
	if file.FileID == "" {
		file.FileID = uuid.New().String()
	}

	// Store metadata
	return repo.metadata.Create(file)
}

// StoreEncryptedData implements the FileRepository.StoreEncryptedData method
func (repo *fileRepositoryImpl) StoreEncryptedData(fileID string, encryptedData []byte) error {
	// Get file metadata to ensure it exists
	file, err := repo.metadata.Get(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("file not found: %s", fileID)
	}

	// Store the encrypted data in S3 - backend doesn't decode/process the content
	storagePath, err := repo.storage.StoreEncryptedData(file.OwnerID, file.FileID, encryptedData)
	if err != nil {
		return err
	}

	// Update the file storage path and size
	file.StoragePath = storagePath
	file.FileSize = int64(len(encryptedData))

	// Update metadata
	return repo.metadata.Update(file)
}
