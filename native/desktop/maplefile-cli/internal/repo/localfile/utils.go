// monorepo/native/desktop/maplefile-cli/internal/repo/localfile/utils.go
package localfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// generateKey creates a storage key for a file metadata
func (r *localFileRepository) generateKey(id string) string {
	return fmt.Sprintf("%s%s", fileKeyPrefix, id)
}

// getAppDataPath gets the path to store file data
func (r *localFileRepository) getAppDataPath(ctx context.Context) (string, error) {
	appDirPath, err := r.configService.GetAppDirPath(ctx)
	if err != nil {
		return "", errors.NewAppError("failed to get app directory path", err)
	}

	// Create data directory if it doesn't exist
	dataPath := filepath.Join(appDirPath, "data")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return "", errors.NewAppError("failed to create data directory", err)
	}

	return dataPath, nil
}
