// cloud/backend/internal/maplefile/service/file/utils.go
package file

import (
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// Helper function to map a File domain model to a FileResponseDTO
func mapFileToDTO(file *dom_file.File) *FileResponseDTO {
	return &FileResponseDTO{
		ID:                            file.ID,
		CollectionID:                  file.CollectionID,
		OwnerID:                       file.OwnerID,
		EncryptedMetadata:             file.EncryptedMetadata,
		EncryptedFileKey:              file.EncryptedFileKey,
		EncryptionVersion:             file.EncryptionVersion,
		EncryptedHash:                 file.EncryptedHash,
		EncryptedFileSizeInBytes:      file.EncryptedFileSizeInBytes,
		EncryptedThumbnailSizeInBytes: file.EncryptedThumbnailSizeInBytes,
		CreatedAt:                     file.CreatedAt,
		ModifiedAt:                    file.ModifiedAt,
		Version:                       file.Version,
		State:                         file.State,
		TombstoneVersion:              file.TombstoneVersion,
		TombstoneExpiry:               file.TombstoneExpiry,
	}
}
