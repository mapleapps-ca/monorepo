// monorepo/native/desktop/maplefile-cli/internal/repo/localfile/datacontent.go
package localfile

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"go.uber.org/zap"
)

// ImportFile imports an existing file into the system
func (r *localFileRepository) ImportFile(ctx context.Context, filePath string, file *localfile.LocalFile) error {
	r.logger.Debug("Importing file into local storage",
		zap.String("sourceFilePath", filePath),
		zap.String("fileID", file.ID.Hex()),
		zap.String("storageMode", file.StorageMode))

	// Get the app data path
	dataPath, err := r.getAppDataPath(ctx)
	if err != nil {
		return err
	}

	// Define paths for both versions
	encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
	decryptedFileName := fmt.Sprintf("%s_decrypted.bin", file.ID.Hex())

	encryptedPath := filepath.Join(dataPath, encryptedFileName)
	decryptedPath := filepath.Join(dataPath, decryptedFileName)

	// Ensure the parent directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return errors.NewAppError("failed to create directory for file", err)
	}

	// Handle each possible storage mode
	switch file.StorageMode {
	case localfile.StorageModeEncryptedOnly:
		// Create only encrypted version
		if err := r.createEncryptedFile(filePath, encryptedPath, file); err != nil {
			return err
		}
		file.EncryptedFilePath = encryptedPath
		file.DecryptedFilePath = ""

	case localfile.StorageModeDecryptedOnly:
		// Create only decrypted version
		if err := r.createDecryptedFile(filePath, decryptedPath, file); err != nil {
			return err
		}
		file.DecryptedFilePath = decryptedPath
		file.EncryptedFilePath = ""

	case localfile.StorageModeHybrid:
		// Create both versions
		if err := r.createEncryptedFile(filePath, encryptedPath, file); err != nil {
			return err
		}
		if err := r.createDecryptedFile(filePath, decryptedPath, file); err != nil {
			// If decrypted fails, try to clean up the encrypted file
			os.Remove(encryptedPath)
			return err
		}
		file.EncryptedFilePath = encryptedPath
		file.DecryptedFilePath = decryptedPath

	default:
		return errors.NewAppError("unsupported storage mode", nil)
	}

	// Update file metadata
	file.ModifiedAt = time.Now()
	file.IsModifiedLocally = true

	// Save the updated file metadata
	if err := r.Save(ctx, file); err != nil {
		return errors.NewAppError("failed to save file metadata after import", err)
	}

	r.logger.Info("File imported successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.String("sourceFilePath", filePath),
		zap.String("storageMode", file.StorageMode))
	return nil
}

// Helper method to create encrypted version of file
func (r *localFileRepository) createEncryptedFile(sourcePath, destPath string, file *localfile.LocalFile) error {
	// Read the source file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return errors.NewAppError("failed to read source file", err)
	}

	var encryptedData []byte

	// If encryption is required (should almost always be the case for this method)
	if file.EncryptionVersion != "unencrypted" {
		// Get the encryption key from file.EncryptedFileKey
		// Decrypt it using the collection key
		// Use it to encrypt the file data

		// Since we don't have the actual encryption implementation, we'll simulate:
		// In a real implementation, you would:
		// 1. Get collection key
		// 2. Decrypt file.EncryptedFileKey to get the file key
		// 3. Use file key to encrypt sourceData
		encryptedData = sourceData // Placeholder for actual encryption
	} else {
		// This shouldn't normally happen but handling it just in case
		encryptedData = sourceData
	}

	// Write the encrypted data to the destination
	if err := os.WriteFile(destPath, encryptedData, 0644); err != nil {
		return errors.NewAppError("failed to write encrypted file", err)
	}

	file.EncryptedSize = int64(len(encryptedData))
	return nil
}

// Helper method to create decrypted version of file
func (r *localFileRepository) createDecryptedFile(sourcePath, destPath string, file *localfile.LocalFile) error {
	// For decrypted files, we simply copy the source file
	// In a real implementation, if the source is already encrypted,
	// you would decrypt it first

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return errors.NewAppError("failed to open source file", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return errors.NewAppError("failed to create destination file", err)
	}
	defer destFile.Close()

	written, err := io.Copy(destFile, sourceFile)
	if err != nil {
		return errors.NewAppError("failed to copy file data", err)
	}

	file.OriginalSize = written
	return nil
}

// SaveFileData saves file data to the local filesystem
// This method needs to be updated to be compatible with our new storage model
func (r *localFileRepository) SaveFileData(ctx context.Context, file *localfile.LocalFile, data []byte) error {
	r.logger.Debug("Saving file data to local filesystem",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)),
		zap.String("storageMode", file.StorageMode))

	// Get the app data path
	dataPath, err := r.getAppDataPath(ctx)
	if err != nil {
		return err
	}

	// Behavior depends on storage mode
	switch file.StorageMode {
	case localfile.StorageModeEncryptedOnly:
		// Create only encrypted version
		encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
		encryptedPath := filepath.Join(dataPath, encryptedFileName)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(encryptedPath), 0755); err != nil {
			return errors.NewAppError("failed to create directory for file", err)
		}

		// In a real implementation, encrypt the data here
		// For simplicity, we're just writing it as-is
		if err := os.WriteFile(encryptedPath, data, 0644); err != nil {
			return errors.NewAppError("failed to write encrypted file", err)
		}

		file.EncryptedFilePath = encryptedPath
		file.EncryptedSize = int64(len(data))

	case localfile.StorageModeDecryptedOnly:
		// Create only decrypted version
		decryptedFileName := fmt.Sprintf("%s_decrypted.bin", file.ID.Hex())
		decryptedPath := filepath.Join(dataPath, decryptedFileName)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(decryptedPath), 0755); err != nil {
			return errors.NewAppError("failed to create directory for file", err)
		}

		// Write the data as-is
		if err := os.WriteFile(decryptedPath, data, 0644); err != nil {
			return errors.NewAppError("failed to write decrypted file", err)
		}

		file.DecryptedFilePath = decryptedPath
		file.OriginalSize = int64(len(data))

	case localfile.StorageModeHybrid:
		// Create both versions
		encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
		decryptedFileName := fmt.Sprintf("%s_decrypted.bin", file.ID.Hex())

		encryptedPath := filepath.Join(dataPath, encryptedFileName)
		decryptedPath := filepath.Join(dataPath, decryptedFileName)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(encryptedPath), 0755); err != nil {
			return errors.NewAppError("failed to create directory for file", err)
		}

		// In a real implementation, encrypt for one file and leave the other unencrypted
		// For simplicity, we're writing the same data to both
		if err := os.WriteFile(encryptedPath, data, 0644); err != nil {
			return errors.NewAppError("failed to write encrypted file", err)
		}

		if err := os.WriteFile(decryptedPath, data, 0644); err != nil {
			// If decrypted version fails, clean up encrypted version
			os.Remove(encryptedPath)
			return errors.NewAppError("failed to write decrypted file", err)
		}

		file.EncryptedFilePath = encryptedPath
		file.DecryptedFilePath = decryptedPath
		file.EncryptedSize = int64(len(data))
		file.OriginalSize = int64(len(data))

	default:
		return errors.NewAppError("unsupported storage mode", nil)
	}

	// Update the file metadata
	file.ModifiedAt = time.Now()
	file.IsModifiedLocally = true

	// Save the updated file metadata
	if err := r.Save(ctx, file); err != nil {
		return errors.NewAppError("failed to update file metadata after saving data", err)
	}

	r.logger.Info("File data saved successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.String("storageMode", file.StorageMode))
	return nil
}

// LoadFileData loads file data from the local filesystem
func (r *localFileRepository) LoadFileData(ctx context.Context, file *localfile.LocalFile) ([]byte, error) {
	r.logger.Debug("Loading file data from local filesystem",
		zap.String("fileID", file.ID.Hex()),
		zap.String("storageMode", file.StorageMode))

	// Determine which path to use based on storage mode
	var filePath string

	// Try decrypted path first (if available) as it's more convenient
	if file.DecryptedFilePath != "" {
		filePath = file.DecryptedFilePath
	} else if file.EncryptedFilePath != "" {
		filePath = file.EncryptedFilePath
		// In a real implementation, you'd need to decrypt this data before returning it
	} else {
		return nil, errors.NewAppError("file has no local path", nil)
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.NewAppError("file data not found on local filesystem", err)
	}

	// Read the file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read file data", err)
	}

	r.logger.Debug("File data loaded successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))
	return data, nil
}
