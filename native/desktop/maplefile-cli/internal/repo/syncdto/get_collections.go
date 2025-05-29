// native/desktop/maplefile-cli/internal/repo/syncdto/impl.go
package syncdto

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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

func (r *syncRepository) GetCollectionSyncDataFromCloud(ctx context.Context, cursor *syncdto.SyncCursorDTO, limit int64) (*syncdto.CollectionSyncResponseDTO, error) {
	r.logger.Debug("🔍 Getting collection syncdto data from cloud",
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

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

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/sync/collections", serverURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error("❌ Failed to parse base URL", zap.String("url", baseURL), zap.Error(err))
		return nil, errors.NewAppError("failed to parse base URL", err)
	}

	// Add query parameters
	queryParams := parsedURL.Query()
	if limit > 0 {
		queryParams.Set("limit", strconv.FormatInt(limit, 10))
	}
	if cursor != nil {
		cursorJSON, err := json.Marshal(cursor)
		if err != nil {
			r.logger.Error("❌ Failed to marshal cursor", zap.Error(err))
			return nil, errors.NewAppError("failed to marshal cursor", err)
		}
		queryParams.Set("cursor", string(cursorJSON))
	}
	parsedURL.RawQuery = queryParams.Encode()

	finalURL := parsedURL.String()
	r.logger.Debug("🔍 Making request to collection syncdto endpoint", zap.String("url", finalURL))

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
	r.logger.Debug("🔍 Executing HTTP request for collection syncdto")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("🔍 Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("⚠️ Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("⚠️ Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	r.logger.Debug("🔍 Parsing HTTP response body into CollectionSyncResponseDTO")
	var response syncdto.CollectionSyncResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into CollectionSyncResponseDTO",
			zap.ByteString("body", body),
			zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Successfully retrieved collection syncdto data from cloud",
		zap.Int("collections_count", len(response.Collections)),
		zap.Bool("has_more", response.HasMore))

	return &response, nil
}
