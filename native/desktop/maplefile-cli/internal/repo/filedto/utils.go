// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/utils.go
package filedto

import (
	"fmt"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// Helper functions for converting between domain and DTO types

// ConvertDomainToDTO converts a domain FileDTO to the repository DTO format
func ConvertDomainToDTO(domainFile *filedto.FileDTO) *filedto.FileDTO {
	// In this case, they're the same type, so just return as-is
	// This function exists for future extensibility
	return domainFile
}

// ConvertDTOToDomain converts a repository DTO to the domain format
func ConvertDTOToDomain(dtoFile *filedto.FileDTO) *filedto.FileDTO {
	// In this case, they're the same type, so just return as-is
	// This function exists for future extensibility
	return dtoFile
}

// ConvertDomainEncryptedFileKey converts domain EncryptedFileKey to DTO EncryptedFileKey
func ConvertDomainEncryptedFileKey(domainKey keys.EncryptedFileKey) filedto.EncryptedFileKey {
	return filedto.EncryptedFileKey{
		Ciphertext: domainKey.Ciphertext,
		Nonce:      domainKey.Nonce,
	}
}

// ConvertDTOEncryptedFileKey converts DTO EncryptedFileKey to domain EncryptedFileKey
func ConvertDTOEncryptedFileKey(dtoKey filedto.EncryptedFileKey) keys.EncryptedFileKey {
	return keys.EncryptedFileKey{
		Ciphertext: dtoKey.Ciphertext,
		Nonce:      dtoKey.Nonce,
		KeyVersion: 1, // Default version
		RotatedAt:  nil,
	}
}

// ValidateCreatePendingFileRequest validates the create pending file request
func ValidateCreatePendingFileRequest(request *filedto.CreatePendingFileRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.CollectionID.IsZero() {
		return fmt.Errorf("collection ID is required")
	}

	if request.EncryptedMetadata == "" {
		return fmt.Errorf("encrypted metadata is required")
	}

	if len(request.EncryptedFileKey.Ciphertext) == 0 {
		return fmt.Errorf("encrypted file key ciphertext is required")
	}

	if len(request.EncryptedFileKey.Nonce) == 0 {
		return fmt.Errorf("encrypted file key nonce is required")
	}

	if request.EncryptionVersion == "" {
		return fmt.Errorf("encryption version is required")
	}

	if request.EncryptedHash == "" {
		return fmt.Errorf("encrypted hash is required")
	}

	if request.ExpectedFileSizeInBytes <= 0 {
		return fmt.Errorf("expected file size must be greater than 0")
	}

	return nil
}

// ValidateCompleteFileUploadRequest validates the complete file upload request
func ValidateCompleteFileUploadRequest(request *filedto.CompleteFileUploadRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Note: All fields in CompleteFileUploadRequest are optional
	// as the server will validate actual sizes against expected sizes
	// and can determine upload status from the storage service

	return nil
}

// IsUploadURLExpired checks if the upload URL has expired
func IsUploadURLExpired(expirationTime time.Time) bool {
	return time.Now().After(expirationTime)
}

// GetTimeUntilExpiration returns the duration until the upload URL expires
func GetTimeUntilExpiration(expirationTime time.Time) time.Duration {
	remaining := time.Until(expirationTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DefaultUploadTimeout returns the default timeout for upload operations
func DefaultUploadTimeout() time.Duration {
	return 5 * time.Minute // 5 minutes for uploads
}

// DefaultDownloadTimeout returns the default timeout for download operations
func DefaultDownloadTimeout() time.Duration {
	return 2 * time.Minute // 2 minutes for downloads
}
