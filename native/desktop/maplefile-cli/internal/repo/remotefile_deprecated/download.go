// internal/repo/remotefile/download.go
package remotefile

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetDownloadURL gets a pre-signed URL for downloading a file by fetching file metadata
func (r *remoteFileRepository) GetDownloadURL(ctx context.Context, fileID primitive.ObjectID) (string, error) {
	// Fetch the file metadata which now includes the download URL
	file, err := r.Fetch(ctx, fileID)
	if err != nil {
		return "", errors.NewAppError("failed to fetch file metadata for download URL", err)
	}

	// Check if download URL is available
	if file.DownloadURL == "" {
		return "", errors.NewAppError("no download URL available for this file", nil)
	}

	// Check if download URL has expired
	if file.DownloadExpiry != "" {
		expiryTime, err := time.Parse(time.RFC3339, file.DownloadExpiry)
		if err != nil {
			r.logger.Warn("Failed to parse download URL expiry time",
				zap.String("fileID", fileID.Hex()),
				zap.String("downloadExpiry", file.DownloadExpiry),
				zap.Error(err))
			// Continue even if we can't parse the expiry time
		} else if time.Now().After(expiryTime) {
			return "", errors.NewAppError("download URL has expired, please fetch file metadata again", nil)
		}
	}

	return file.DownloadURL, nil
}

// DownloadFile downloads file data from the cloud server using the presigned URL
func (r *remoteFileRepository) DownloadFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error) {
	r.logger.Info("Starting file download using presigned URL",
		zap.String("fileID", fileID.Hex()))

	// Get the file metadata with download URL
	file, err := r.Fetch(ctx, fileID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch file metadata for download", err)
	}

	// Check if download URL is available
	if file.DownloadURL == "" {
		return nil, errors.NewAppError("no download URL available for this file", nil)
	}

	// Check if download URL has expired
	if file.DownloadExpiry != "" {
		expiryTime, err := time.Parse(time.RFC3339, file.DownloadExpiry)
		if err != nil {
			r.logger.Warn("Failed to parse download URL expiry time, continuing with download",
				zap.String("fileID", fileID.Hex()),
				zap.String("downloadExpiry", file.DownloadExpiry),
				zap.Error(err))
		} else if time.Now().After(expiryTime) {
			return nil, errors.NewAppError("download URL has expired, please fetch file metadata again", nil)
		}
	}

	r.logger.Debug("Using presigned URL for file download",
		zap.String("fileID", fileID.Hex()),
		zap.String("downloadExpiry", file.DownloadExpiry))

	// Create HTTP request using the presigned URL
	req, err := http.NewRequestWithContext(ctx, "GET", file.DownloadURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create download request", err)
	}

	// Note: No Authorization header needed for presigned URLs

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to download file from presigned URL", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.logger.Error("Failed to download file using presigned URL",
			zap.String("fileID", fileID.Hex()),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.ByteString("responseBody", body))
		return nil, errors.NewAppError(fmt.Sprintf("download failed with status: %s", resp.Status), nil)
	}

	// Read file data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read downloaded file data", err)
	}

	r.logger.Info("Successfully downloaded file using presigned URL",
		zap.String("fileID", fileID.Hex()),
		zap.Int("downloadedBytes", len(data)))

	return data, nil
}
