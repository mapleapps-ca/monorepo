// internal/repo/remotefile/upload.go
package remotefile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// UploadFileByEncryptedID uploads file data using the remote file ID
func (r *remoteFileRepository) UploadFileByRemoteID(ctx context.Context, remoteID primitive.ObjectID, data []byte) error {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address",
			zap.Error(err),
		)
		return errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token for file upload",
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("Starting file upload to backend using encrypted file ID",
		zap.Int("dataSize", len(data)))

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", "encrypted_file.bin")
	if err != nil {
		r.logger.Error("Failed to create form file field",
			zap.Error(err),
		)
		return errors.NewAppError("failed to create form file field", err)
	}

	// Write file data
	if _, err := fileWriter.Write(data); err != nil {
		r.logger.Error("Failed to write file data to form",
			zap.Error(err),
		)
		return errors.NewAppError("failed to write file data to form", err)
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		r.logger.Error("Failed to close multipart writer",
			zap.Error(err),
		)
		return errors.NewAppError("failed to close multipart writer", err)
	}

	// Create HTTP request using encrypted file ID in URL
	uploadURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/data", serverURL, remoteID.Hex())
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &requestBody)
	if err != nil {
		r.logger.Error("Failed to create upload request",
			zap.Error(err),
		)
		return errors.NewAppError("failed to create upload request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to execute upload request",
			zap.Error(err),
		)
		return errors.NewAppError("failed to execute upload request", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read upload response body",
			zap.Error(err),
		)
		return errors.NewAppError("failed to read upload response body", err)
	}

	// Check for success status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		r.logger.Error("Backend returned error status during file upload",
			zap.Int("status", resp.StatusCode),
			zap.String("statusText", resp.Status),
			zap.ByteString("responseBody", body),
		)
		return errors.NewAppError(fmt.Sprintf("backend upload failed with status: %s, body: %s", resp.Status, string(body)), nil)
	}

	// Parse success response
	var response struct {
		Success       bool   `json:"success"`
		Message       string `json:"message"`
		FileObjectKey string `json:"file_object_key"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Warn("Failed to parse upload response, but upload succeeded",
			zap.Error(err),
		)
		// Don't return error since upload succeeded
	}

	r.logger.Info("Successfully uploaded file to backend/S3",
		zap.Int("uploadedBytes", len(data)),
		zap.String("fileObjectKey", response.FileObjectKey))

	return nil
}

// UploadFile uploads file data using MongoDB ObjectID (legacy method, kept for compatibility)
func (r *remoteFileRepository) UploadFileByLocalID(ctx context.Context, fileID primitive.ObjectID, data []byte) error {
	// First, get the file to obtain its remote file
	remotefile, err := r.Fetch(ctx, fileID)
	if err != nil {
		return errors.NewAppError("failed to fetch file to get encrypted file ID", err)
	}

	// Use the encrypted file ID for upload
	return r.UploadFileByRemoteID(ctx, remotefile.ID, data)
}
