// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/list_from_cloud.go
package filedto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// ListFromCloud lists FileDTOs from the cloud service based on the provided filter criteria
func (r *fileDTORepository) ListFromCloud(ctx context.Context, filter filedto.FileFilter) ([]*filedto.FileDTO, error) {
	r.logger.Debug("Listing files from cloud",
		zap.Any("filter", filter))

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token for authentication
	accessToken, err := r.tokenRepo.GetAccessToken(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Build request URL based on filter
	var requestURL string
	if filter.CollectionID != nil {
		// List files by collection
		requestURL = fmt.Sprintf("%s/maplefile/api/v1/collections/%s/files", serverURL, filter.CollectionID.Hex())
	} else {
		// This would require a general file listing endpoint, which doesn't seem to exist in the backend
		// For now, we'll return an error requesting a collection filter
		return nil, errors.NewAppError("collection ID filter is required for listing files", nil)
	}

	// Add query parameters for additional filters
	if filter.State != "" {
		params := url.Values{}
		params.Add("state", filter.State)
		requestURL += "?" + params.Encode()
	}

	r.logger.Debug("Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

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
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var filesResponse struct {
		Files []*filedto.FileDTO `json:"files"`
	}
	if err := json.Unmarshal(body, &filesResponse); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Ensure we return a non-nil slice even if empty
	if filesResponse.Files == nil {
		filesResponse.Files = []*filedto.FileDTO{}
	}

	r.logger.Info("Successfully listed files from cloud",
		zap.Int("count", len(filesResponse.Files)),
		zap.String("collectionID", filter.CollectionID.Hex()))

	return filesResponse.Files, nil
}
