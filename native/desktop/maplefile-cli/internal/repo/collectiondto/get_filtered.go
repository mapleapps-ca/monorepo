// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/get_filtered.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

func (r *collectionDTORepository) GetFilteredCollectionsFromCloud(ctx context.Context, request *collectiondto.GetFilteredCollectionsRequest) (*collectiondto.GetFilteredCollectionsResponse, error) {
	r.logger.Debug("🔍 Getting filtered collections from cloud",
		zap.Bool("include_owned", request.IncludeOwned),
		zap.Bool("include_shared", request.IncludeShared))

	// Get access token
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Validate request
	if !request.IncludeOwned && !request.IncludeShared {
		return nil, errors.NewAppError("at least one filter option (include_owned or include_shared) must be enabled", nil)
	}

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/collections/filtered", serverURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error("❌ Failed to parse base URL", zap.String("url", baseURL), zap.Error(err))
		return nil, errors.NewAppError("failed to parse base URL", err)
	}

	// Add query parameters
	queryParams := parsedURL.Query()
	queryParams.Set("include_owned", strconv.FormatBool(request.IncludeOwned))
	queryParams.Set("include_shared", strconv.FormatBool(request.IncludeShared))
	parsedURL.RawQuery = queryParams.Encode()

	finalURL := parsedURL.String()
	r.logger.Debug("➡️ Making request to filtered collections endpoint", zap.String("url", finalURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request for filtered collections")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("➡️ Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("🚨 Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("🚨 Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	r.logger.Debug("➡️ Parsing HTTP response body into GetFilteredCollectionsResponse")
	var response collectiondto.GetFilteredCollectionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into GetFilteredCollectionsResponse",
			zap.ByteString("body", body),
			zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Successfully retrieved filtered collections from cloud",
		zap.Int("owned_count", len(response.OwnedCollections)),
		zap.Int("shared_count", len(response.SharedCollections)),
		zap.Int("total_count", response.TotalCount))

	return &response, nil
}
