// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/upload_file_to_cloud.go
package filedto

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// UploadFileToCloud uploads the actual file content to cloud storage using presigned URL
func (r *fileDTORepository) UploadFileToCloud(ctx context.Context, presignedURL string, fileData []byte) error {
	r.logger.Debug("Uploading file content to cloud storage",
		zap.String("presignedURL", presignedURL),
		zap.Int("dataSize", len(fileData)))

	if presignedURL == "" {
		return errors.NewAppError("presigned URL is required", nil)
	}

	if len(fileData) == 0 {
		return errors.NewAppError("file data is required", nil)
	}

	// Create HTTP PUT request to the presigned URL
	req, err := http.NewRequestWithContext(ctx, "PUT", presignedURL, bytes.NewReader(fileData))
	if err != nil {
		return errors.NewAppError("failed to create HTTP request for file upload", err)
	}

	// Set content type for binary data
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(fileData))

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.NewAppError("failed to upload file to cloud storage", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Warn("Failed to read upload response body", zap.Error(err))
	}

	// Check for successful upload
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.NewAppError(fmt.Sprintf("file upload failed with status %d: %s", resp.StatusCode, string(body)), nil)
	}

	r.logger.Info("Successfully uploaded file content to cloud storage",
		zap.Int("dataSize", len(fileData)),
		zap.Int("statusCode", resp.StatusCode))

	return nil
}

// UploadThumbnailToCloud uploads thumbnail content to cloud storage using presigned URL
func (r *fileDTORepository) UploadThumbnailToCloud(ctx context.Context, presignedURL string, thumbnailData []byte) error {
	r.logger.Debug("Uploading thumbnail content to cloud storage",
		zap.String("presignedURL", presignedURL),
		zap.Int("dataSize", len(thumbnailData)))

	if presignedURL == "" {
		return errors.NewAppError("presigned thumbnail URL is required", nil)
	}

	if len(thumbnailData) == 0 {
		r.logger.Debug("No thumbnail data provided, skipping thumbnail upload")
		return nil
	}

	// Create HTTP PUT request to the presigned URL
	req, err := http.NewRequestWithContext(ctx, "PUT", presignedURL, bytes.NewReader(thumbnailData))
	if err != nil {
		return errors.NewAppError("failed to create HTTP request for thumbnail upload", err)
	}

	// Set content type for image data
	req.Header.Set("Content-Type", "image/jpeg") // Default to JPEG, could be made configurable
	req.ContentLength = int64(len(thumbnailData))

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.NewAppError("failed to upload thumbnail to cloud storage", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Warn("Failed to read thumbnail upload response body", zap.Error(err))
	}

	// Check for successful upload
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.NewAppError(fmt.Sprintf("thumbnail upload failed with status %d: %s", resp.StatusCode, string(body)), nil)
	}

	r.logger.Info("Successfully uploaded thumbnail content to cloud storage",
		zap.Int("dataSize", len(thumbnailData)),
		zap.Int("statusCode", resp.StatusCode))

	return nil
}
