// internal/repo/remotefile/fetch.go
package remotefile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDownloadURL gets a pre-signed URL for downloading a file
func (r *remoteFileRepository) GetDownloadURL(ctx context.Context, fileID primitive.ObjectID) (string, error) {
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
	downloadURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/download-url", serverURL, fileID.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
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

// DownloadFile downloads file data from the remote server
func (r *remoteFileRepository) DownloadFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error) {
	// Get pre-signed download URL
	downloadURL, err := r.GetDownloadURL(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to download file", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s, body: %s", resp.Status, string(body)), nil)
	}

	// Read file data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read file data", err)
	}

	return data, nil
}
