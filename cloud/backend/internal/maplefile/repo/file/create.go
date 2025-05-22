// cloud/backend/internal/maplefile/repo/file/create.go
package file

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// Create implements the FileRepository.Create method
func (repo *fileRepositoryImpl) Create(file *dom_file.File) error {
	// Generate ID if not provided
	if file.ID.IsZero() {
		file.ID = primitive.NewObjectID()
	}

	// If EncryptedFileID is not set, generate one - ideally this comes from client
	if file.EncryptedFileID == "" {
		file.EncryptedFileID = primitive.NewObjectID().Hex()
	}

	// Store metadata
	return repo.metadata.Create(file)
}

// CreateMany implements the FileRepository.CreateMany method
func (repo *fileRepositoryImpl) CreateMany(files []*dom_file.File) error {
	// Pre-process files to ensure they have IDs
	for _, file := range files {
		if file.ID.IsZero() {
			file.ID = primitive.NewObjectID()
		}
		if file.EncryptedFileID == "" {
			file.EncryptedFileID = primitive.NewObjectID().Hex()
		}
	}

	// Store metadata for all files
	return repo.metadata.CreateMany(files)
}

// StoreEncryptedData implements the FileRepository.StoreEncryptedData method
func (repo *fileRepositoryImpl) StoreEncryptedData(fileID string, encryptedData []byte) error {
	// Get file metadata to ensure it exists
	file, err := repo.metadata.GetByEncryptedFileID(fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("file not found: %s", fileID)
	}

	// Store the encrypted data in S3 - backend doesn't decode/process the content
	storagePath, err := repo.storage.StoreEncryptedData(file.OwnerID.Hex(), fileID, encryptedData)
	if err != nil {
		return err
	}

	// Update the file storage path and size
	file.FileObjectKey = storagePath
	file.FileSize = int64(len(encryptedData))

	// Update metadata
	return repo.metadata.Update(file)
}
