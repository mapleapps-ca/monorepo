// internal/repo/remotefile/upload.go
package remotefile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetUploadURL gets a pre-signed URL for uploading a file
func (r *remoteFileRepository) GetUploadURL(ctx context.Context, fileID primitive.ObjectID) (string, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address",
			zap.Error(err),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token for upload URL",
			zap.Error(err),
			zap.Any("fileID", fileID),
		)
		return "", err
	}

	// Create HTTP request
	uploadURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/upload-url", serverURL, fileID.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", uploadURL, nil)
	if err != nil {
		r.logger.Error("Failed making HTTP request for upload URL",
			zap.Error(err),
			zap.Any("upload_url", uploadURL),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to connect to server for upload URL",
			zap.Error(err),
			zap.Any("upload_url", uploadURL),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read response body for upload URL",
			zap.Error(err),
			zap.Any("upload_url", uploadURL),
			zap.Int("status", resp.StatusCode),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("Server returned error status for upload URL",
			zap.Int("status", resp.StatusCode),
			zap.String("statusText", resp.Status),
			zap.ByteString("responseBody", body),
			zap.Any("upload_url", uploadURL),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body for upload URL",
			zap.Error(err),
			zap.ByteString("responseBody", body),
			zap.Any("upload_url", uploadURL),
			zap.Int("status", resp.StatusCode),
			zap.Any("fileID", fileID),
		)
		return "", errors.NewAppError("failed to parse response", err)
	}

	return response.URL, nil
}

// UploadFile uploads file data directly to S3 using presigned URL
func (r *remoteFileRepository) UploadFile(ctx context.Context, fileID primitive.ObjectID, data []byte) error {
	// Step 1: Get pre-signed upload URL from backend
	uploadURL, err := r.GetUploadURL(ctx, fileID)
	if err != nil {
		// Error is already logged within GetUploadURL
		return err
	}

	r.logger.Info("Starting S3 upload using presigned URL",
		zap.String("fileID", fileID.Hex()),
		zap.Int("dataSize", len(data)),
		zap.String("uploadURL", uploadURL[:min(len(uploadURL), 100)]+"...")) // Log first 100 chars for security

	// Step 2: Upload directly to S3 using presigned URL
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewBuffer(data))
	if err != nil {
		r.logger.Error("Failed to create S3 upload request",
			zap.Error(err),
			zap.String("fileID", fileID.Hex()),
		)
		return errors.NewAppError("failed to create S3 upload request", err)
	}

	// Set appropriate headers for S3
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(data))

	// Execute the S3 upload
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to upload file to S3",
			zap.Error(err),
			zap.String("fileID", fileID.Hex()),
		)
		return errors.NewAppError("failed to upload file to S3", err)
	}
	defer resp.Body.Close()

	// Check for S3 upload success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			r.logger.Error("S3 returned error status during upload, failed to read body",
				zap.Error(readErr),
				zap.Int("status", resp.StatusCode),
				zap.String("statusText", resp.Status),
				zap.String("fileID", fileID.Hex()),
			)
			return errors.NewAppError(fmt.Sprintf("S3 upload failed with status: %s, failed to read body", resp.Status), nil)
		}
		r.logger.Error("S3 returned error status during file upload",
			zap.Int("status", resp.StatusCode),
			zap.String("statusText", resp.Status),
			zap.ByteString("responseBody", body),
			zap.String("fileID", fileID.Hex()),
		)
		return errors.NewAppError(fmt.Sprintf("S3 upload failed with status: %s, body: %s", resp.Status, string(body)), nil)
	}

	r.logger.Info("Successfully uploaded file to S3",
		zap.String("fileID", fileID.Hex()),
		zap.Int("uploadedBytes", len(data)))

	return nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
