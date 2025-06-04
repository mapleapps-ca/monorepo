// internal/service/filesyncer/utils.go
package filesyncer

import (
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// mapFileDTOToDomain maps a FileDTO to a File domain object
func mapFileDTOToDomain(dto *filedto.FileDTO) *dom_file.File {
	if dto == nil {
		return nil
	}

	state := dto.State
	if state == "" {
		state = dom_file.FileStateActive // Default to active
	}

	return &dom_file.File{
		ID:                     dto.ID,
		CollectionID:           dto.CollectionID,
		OwnerID:                dto.OwnerID,
		EncryptedMetadata:      dto.EncryptedMetadata,
		EncryptedFileKey:       dto.EncryptedFileKey,
		EncryptionVersion:      dto.EncryptionVersion,
		EncryptedHash:          dto.EncryptedHash,
		EncryptedFileSize:      dto.EncryptedFileSizeInBytes,
		EncryptedThumbnailSize: dto.EncryptedThumbnailSizeInBytes,
		Name:                   "[Encrypted]",              // Will be handled later in the execution flow
		EncryptedFilePath:      "...",                      // Will be handled later in the execution flow
		Metadata:               nil,                        // Will be handled later in the execution flow
		MimeType:               "application/octet-stream", // Will be handled later in the execution flow
		FilePath:               "...",                      // Will be handled later in the execution flow
		FileSize:               0,                          // Will be handled later in the execution flow
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,
		State:                  state,
		SyncStatus:             dom_file.SyncStatusCloudOnly,
		StorageMode:            dom_file.StorageModeEncryptedOnly,
	}
}
