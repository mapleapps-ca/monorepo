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
)

// GetUploadURL gets a pre-signed URL for uploading a file
func (r *remoteFileRepository) GetUploadURL(ctx context.Context, fileID primitive.ObjectID) (string, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return "", errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	uploadURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/upload-url", serverURL, fileID.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", uploadURL, nil)
	if err != nil {
		return "", errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		return "", errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", errors.NewAppError("failed to parse response", err)
	}

	return response.URL, nil
}

// UploadFile uploads file data to the remote server
func (r *remoteFileRepository) UploadFile(ctx context.Context, fileID primitive.ObjectID, data []byte) error {
	// Get pre-signed upload URL
	uploadURL, err := r.GetUploadURL(ctx, fileID)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewBuffer(data))
	if err != nil {
		return errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.NewAppError("failed to upload file", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAppError(fmt.Sprintf("server returned error status: %s, body: %s", resp.Status, string(body)), nil)
	}

	return nil
}
