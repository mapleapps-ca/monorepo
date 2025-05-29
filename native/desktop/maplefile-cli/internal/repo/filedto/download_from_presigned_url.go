// native/desktop/maplefile-cli/internal/repo/filedto/download_from_presigned_url.go
package filedto

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// DownloadFileViaPresignedURLFromCloud downloads file content from a presigned URL
func (r *fileDTORepository) DownloadFileViaPresignedURLFromCloud(ctx context.Context, presignedURL string) ([]byte, error) {
	r.logger.Debug("⬇️ Downloading file from presigned URL",
		zap.String("presignedURL", presignedURL))

	if presignedURL == "" {
		return nil, errors.NewAppError("presigned URL is required", nil)
	}

	// Create HTTP GET request to the presigned URL
	req, err := http.NewRequestWithContext(ctx, "GET", presignedURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request for file download", err)
	}

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to download file from presigned URL", err)
	}
	defer resp.Body.Close()

	// Check for successful download
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read response body for error details
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			r.logger.Warn("⚠️ Failed to read download error response body", zap.Error(err))
		}
		return nil, errors.NewAppError(fmt.Sprintf("file download failed with status %d: %s", resp.StatusCode, string(body)), nil)
	}

	// Read the file content
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read downloaded file data", err)
	}

	r.logger.Info("✅ Successfully downloaded file from presigned URL",
		zap.Int("dataSize", len(fileData)),
		zap.Int("statusCode", resp.StatusCode))

	return fileData, nil
}

// DownloadThumbnailViaPresignedURLFromCloud downloads thumbnail content from a presigned URL
func (r *fileDTORepository) DownloadThumbnailViaPresignedURLFromCloud(ctx context.Context, presignedURL string) ([]byte, error) {
	r.logger.Debug("🖼️ Downloading thumbnail from presigned URL",
		zap.String("presignedURL", presignedURL))

	if presignedURL == "" {
		r.logger.Debug("ℹ️ No presigned thumbnail URL provided, skipping thumbnail download")
		return nil, nil
	}

	// Use the same logic as file download
	return r.DownloadFileViaPresignedURLFromCloud(ctx, presignedURL)
}
