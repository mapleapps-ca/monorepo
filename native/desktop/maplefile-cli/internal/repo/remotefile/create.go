// internal/repo/remotefile/create.go
package remotefile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// Create creates a new file in the remote server
func (r *remoteFileRepository) Create(ctx context.Context, request *remotefile.RemoteCreateFileRequest) (*remotefile.RemoteFileResponse, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	createURL := fmt.Sprintf("%s/maplefile/api/v1/files", serverURL)
	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response remotefile.RemoteFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	return &response, nil
}
