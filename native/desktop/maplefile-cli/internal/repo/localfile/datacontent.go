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

// SaveFileData saves file data to the local filesystem
func (r *localFileRepository) SaveFileData(ctx context.Context, file *localfile.LocalFile, data []byte) error {
	r.logger.Debug("Saving file data to local filesystem",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))

	// Get the app data path
	dataPath, err := r.getAppDataPath(ctx)
	if err != nil {
		return err
	}

	// Create a unique filename based on the file ID
	filename := fmt.Sprintf("%s.bin", file.ID.Hex())
	filePath := filepath.Join(dataPath, filename)

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.NewAppError("failed to create directory for file", err)
	}

	// Write the file data
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.NewAppError("failed to write file data", err)
	}

	// Update the file metadata with the local path
	file.LocalFilePath = filePath
	file.EncryptedSize = int64(len(data))
	file.ModifiedAt = time.Now()
	file.IsModifiedLocally = true

	// Save the updated file metadata
	if err := r.Save(ctx, file); err != nil {
		return errors.NewAppError("failed to update file metadata after saving data", err)
	}

	r.logger.Info("File data saved successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.String("localFilePath", filePath))
	return nil
}

// LoadFileData loads file data from the local filesystem
func (r *localFileRepository) LoadFileData(ctx context.Context, file *localfile.LocalFile) ([]byte, error) {
	r.logger.Debug("Loading file data from local filesystem",
		zap.String("fileID", file.ID.Hex()),
		zap.String("localFilePath", file.LocalFilePath))

	// Check if the file has a local path
	if file.LocalFilePath == "" {
		return nil, errors.NewAppError("file has no local path", nil)
	}

	// Check if the file exists
	if _, err := os.Stat(file.LocalFilePath); os.IsNotExist(err) {
		return nil, errors.NewAppError("file data not found on local filesystem", err)
	}

	// Read the file data
	data, err := os.ReadFile(file.LocalFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read file data", err)
	}

	r.logger.Debug("File data loaded successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("dataSize", len(data)))
	return data, nil
}

// ImportFile imports an existing file into the system
func (r *localFileRepository) ImportFile(ctx context.Context, filePath string, file *localfile.LocalFile) error {
	r.logger.Debug("Importing file into local storage",
		zap.String("sourceFilePath", filePath),
		zap.String("fileID", file.ID.Hex()),
		zap.String("fileState", file.LocalFileState))

	// Get the app data path
	dataPath, err := r.getAppDataPath(ctx)
	if err != nil {
		return err
	}

	// Create a unique filename based on the file ID
	filename := fmt.Sprintf("%s.bin", file.ID.Hex())
	destinationPath := filepath.Join(dataPath, filename)

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return errors.NewAppError("failed to create directory for file", err)
	}

	// Different handling based on file state
	if file.LocalFileState == localfile.LocalFileStateLocalAndEncrypted {
		// For encrypted files, we'd perform encryption here
		// Open source file
		sourceFile, err := os.Open(filePath)
		if err != nil {
			return errors.NewAppError("failed to open source file", err)
		}
		defer sourceFile.Close()

		// Create destination file
		destFile, err := os.Create(destinationPath)
		if err != nil {
			return errors.NewAppError("failed to create destination file", err)
		}
		defer destFile.Close()

		// In a real implementation, we would:
		// 1. Read the source file data
		// 2. Encrypt it using the file key
		// 3. Write the encrypted data to the destination file

		// For this simplified implementation, we'll just copy the file
		written, err := io.Copy(destFile, sourceFile)
		if err != nil {
			return errors.NewAppError("failed to copy file data", err)
		}

		// Update the file metadata
		file.LocalFilePath = destinationPath
		file.EncryptedSize = written
		file.ModifiedAt = time.Now()
		file.IsModifiedLocally = true
	} else if file.LocalFileState == localfile.LocalFileStateLocalAndDecrypted {
		// For unencrypted files, we just copy the file as-is
		sourceFile, err := os.Open(filePath)
		if err != nil {
			return errors.NewAppError("failed to open source file", err)
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destinationPath)
		if err != nil {
			return errors.NewAppError("failed to create destination file", err)
		}
		defer destFile.Close()

		written, err := io.Copy(destFile, sourceFile)
		if err != nil {
			return errors.NewAppError("failed to copy file data", err)
		}

		// Update the file metadata
		file.LocalFilePath = destinationPath
		file.EncryptedSize = written // Even though not encrypted, we use this field
		file.ModifiedAt = time.Now()
		file.IsModifiedLocally = true
	} else {
		return errors.NewAppError("unsupported file state", nil)
	}

	// Save the file metadata
	if err := r.Save(ctx, file); err != nil {
		return errors.NewAppError("failed to save file metadata after import", err)
	}

	r.logger.Info("File imported successfully",
		zap.String("fileID", file.ID.Hex()),
		zap.String("sourceFilePath", filePath),
		zap.String("destinationPath", destinationPath),
		zap.String("fileState", file.LocalFileState))
	return nil
}
