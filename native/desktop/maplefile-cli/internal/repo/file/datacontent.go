// monorepo/native/desktop/maplefile-cli/internal/repo/file/datacontent.go
package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// ImportFile imports an existing file into the system
func (r *fileRepository) ImportFile(ctx context.Context, filePath string, file *dom_file.Collection) error {
	r.logger.Debug("Importing file into local storage",
		zap.String("sourceFilePath", filePath),
		zap.String("fileID", file.ID.Hex()),
		zap.String("storageMode", file.StorageMode))

	// Get the app data path
	dataPath, err := r.getAppDataPath(ctx)
	if err != nil {
		return err
	}

	// Determine the appropriate file extension using only the file path
	fileExt := r.determineFileExtension(filePath)

	// Define paths for both versions with appropriate extensions
	encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
	decryptedFileName := fmt.Sprintf("%s_decrypted%s", file.ID.Hex(), fileExt)

	encryptedPath := filepath.Join(dataPath, encryptedFileName)
	decryptedPath := filepath.Join(dataPath, decryptedFileName)

	// Ensure the parent directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return errors.NewAppError("failed to create directory for file", err)
	}

	// Handle each possible storage mode
	switch file.StorageMode {
	case dom_file.StorageModeEncryptedOnly:
		// Create only encrypted version
		if err := r.createEncryptedFile(filePath, encryptedPath, file); err != nil {
			return err
		}
		file.EncryptedFilePath = encryptedPath
		file.DecryptedFilePath = ""

	case dom_file.StorageModeDecryptedOnly:
		// Create only decrypted version
		if err := r.createDecryptedFile(filePath, decryptedPath, file); err != nil {
			return err
		}
		file.DecryptedFilePath = decryptedPath
		file.EncryptedFilePath = ""

	case dom_file.StorageModeHybrid:
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

// determineFileExtension determines the appropriate file extension for a file
// by only using the file path.
func (r *fileRepository) determineFileExtension(filePath string) string {
	// Only try to get extension from original file path
	return filepath.Ext(filePath)
}

// Helper method to create encrypted version of file
func (r *fileRepository) createEncryptedFile(sourcePath, destPath string, file *file.Collection) error {
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

	file.EncryptedFileSize = int64(len(encryptedData))
	return nil
}

// Helper method to create decrypted version of file
func (r *fileRepository) createDecryptedFile(sourcePath, destPath string, file *file.Collection) error {
	// For decrypted files, we simply copy the source file as-is
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

	file.DecryptedFileSize = written
	return nil
}

// saveEncryptedFileDataInternal handles saving data for StorageModeEncryptedOnly.
// It is called by SaveFileData when the mode is EncryptedOnly.
func (r *fileRepository) SaveEncryptedFileDataInternal(ctx context.Context, dataPath string, file *dom_file.Collection, data []byte) error {
	r.logger.Debug("Saving encrypted file data internally",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))

	// For encrypted files, use .bin extension
	encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
	encryptedPath := filepath.Join(dataPath, encryptedFileName)

	// In a real implementation, encrypt the data here before writing
	// For simplicity, we're just writing the provided data as-is
	if err := os.WriteFile(encryptedPath, data, 0644); err != nil {
		return errors.NewAppError("failed to write encrypted file data", err)
	}

	file.EncryptedFilePath = encryptedPath
	file.EncryptedFileSize = int64(len(data))
	file.DecryptedFilePath = "" // Clear decrypted path
	file.DecryptedFileSize = 0

	r.logger.Debug("Encrypted file data saved successfully internally",
		zap.String("fileID", file.ID.Hex()),
		zap.String("encryptedPath", encryptedPath))

	return nil
}

// saveDecryptedFileDataInternal handles saving data for StorageModeDecryptedOnly.
// It is called by SaveFileData when the mode is DecryptedOnly.
func (r *fileRepository) SaveDecryptedFileDataInternal(ctx context.Context, dataPath string, file *dom_file.Collection, data []byte) error {
	r.logger.Debug("Saving decrypted file data internally",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))

	// For decrypted files, preserve the original extension from DecryptedName
	fileExt := filepath.Ext(file.DecryptedName)
	if fileExt == "" {
		fileExt = ".bin" // Fallback if no extension is found
	}
	decryptedFileName := fmt.Sprintf("%s_decrypted%s", file.ID.Hex(), fileExt)
	decryptedPath := filepath.Join(dataPath, decryptedFileName)

	if err := os.WriteFile(decryptedPath, data, 0644); err != nil {
		return errors.NewAppError("failed to write decrypted file data", err)
	}

	file.DecryptedFilePath = decryptedPath
	file.DecryptedFileSize = int64(len(data))
	file.EncryptedFilePath = "" // Clear encrypted path
	file.EncryptedFileSize = 0

	r.logger.Debug("Decrypted file data saved successfully internally",
		zap.String("fileID", file.ID.Hex()),
		zap.String("decryptedPath", decryptedPath))

	return nil
}

// saveHybridFileDataInternal handles saving data for StorageModeHybrid.
// It is called by SaveFileData when the mode is Hybrid.
func (r *fileRepository) SaveHybridFileDataInternal(ctx context.Context, dataPath string, file *dom_file.Collection, data []byte) error {
	r.logger.Debug("Saving hybrid file data internally",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))

	// For encrypted files, use .bin extension
	encryptedFileName := fmt.Sprintf("%s_encrypted.bin", file.ID.Hex())
	encryptedPath := filepath.Join(dataPath, encryptedFileName)

	// For decrypted files, preserve the original extension from DecryptedName
	fileExt := filepath.Ext(file.DecryptedName)
	if fileExt == "" {
		fileExt = ".bin" // Fallback if no extension is found
	}
	decryptedFileName := fmt.Sprintf("%s_decrypted%s", file.ID.Hex(), fileExt)
	decryptedPath := filepath.Join(dataPath, decryptedFileName)

	// --- START NOTE ---
	// In a real implementation, you would encrypt `data` for the encrypted file
	// and use the raw `data` (or a different version) for the decrypted file.
	// Following the original code's behavior, we write the *same* input `data` to both files for simplicity.
	// The recorded file sizes (EncryptedFileSize, DecryptedFileSize) should reflect the *actual* sizes
	// of the data written to each file after potential encryption/decryption.
	// --- END NOTE ---

	// Write the encrypted version (using input data as placeholder for encrypted data)
	if err := os.WriteFile(encryptedPath, data, 0644); err != nil {
		return errors.NewAppError("failed to write encrypted file in hybrid mode", err)
	}
	file.EncryptedFilePath = encryptedPath
	file.EncryptedFileSize = int64(len(data)) // Placeholder: Should be actual encrypted size

	// Write the decrypted version (using input data)
	if err := os.WriteFile(decryptedPath, data, 0644); err != nil {
		// If decrypted version fails, clean up encrypted version
		os.Remove(encryptedPath)
		file.EncryptedFilePath = "" // Clear path if cleanup successful or attempted
		file.EncryptedFileSize = 0
		return errors.NewAppError("failed to write decrypted file in hybrid mode", err)
	}
	file.DecryptedFilePath = decryptedPath
	file.DecryptedFileSize = int64(len(data)) // Placeholder: Should be actual decrypted size

	r.logger.Debug("Hybrid file data saved successfully internally",
		zap.String("fileID", file.ID.Hex()),
		zap.String("encryptedPath", encryptedPath),
		zap.String("decryptedPath", decryptedPath))

	return nil
}

// LoadFileData loads file data from the local filesystem
func (r *fileRepository) LoadFileData(ctx context.Context, file *dom_file.Collection) ([]byte, error) {
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

// LoadEncryptedFileData loads encrypted file data from the local filesystem
func (r *fileRepository) LoadDecryptedFileDataAtFilePath(ctx context.Context, decryptedFilePath string) ([]byte, error) {
	r.logger.Debug("Loading decrypted file data from local filesystem",
		zap.String("decryptedFilePath", decryptedFilePath))

	// Check if the file exists
	if _, err := os.Stat(decryptedFilePath); os.IsNotExist(err) {
		return nil, errors.NewAppError("decrypted file data not found on local filesystem", err)
	}

	// Read the file data
	data, err := os.ReadFile(decryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("decrypted failed to read file data", err)
	}

	r.logger.Debug("Decrypted file data loaded successfully",
		zap.String("decryptedFilePath", decryptedFilePath),
		zap.Int("dataSize", len(data)))
	return data, nil
}

// LoadEncryptedFileData loads encrypted file data from the local filesystem
func (r *fileRepository) LoadEncryptedFileDataAtFilePath(ctx context.Context, encryptedFilePath string) ([]byte, error) {
	r.logger.Debug("Loading encrypted file data from local filesystem",
		zap.String("encryptedFilePath", encryptedFilePath))

	// Check if the file exists
	if _, err := os.Stat(encryptedFilePath); os.IsNotExist(err) {
		return nil, errors.NewAppError("encrypted file data not found on local filesystem", err)
	}

	// Read the file data
	data, err := os.ReadFile(encryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("encrypted failed to read file data", err)
	}

	r.logger.Debug("Encrypted file data loaded successfully",
		zap.String("encryptedFilePath", encryptedFilePath),
		zap.Int("dataSize", len(data)))
	return data, nil
}
