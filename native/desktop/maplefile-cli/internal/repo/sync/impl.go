// native/desktop/maplefile-cli/internal/repo/sync/impl.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/sync"
)

// syncRepository implements the sync.SyncRepository interface
type syncRepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository auth.TokenRepository
	httpClient      *http.Client
}

// NewSyncRepository creates a new repository for sync operations
func NewSyncRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
) sync.SyncRepository {
	return &syncRepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *syncRepository) GetCollectionSyncData(ctx context.Context, cursor *sync.SyncCursor, limit int64) (*sync.CollectionSyncResponse, error) {
	r.logger.Debug("Getting collection sync data from cloud",
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/sync/collections", serverURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error("Failed to parse base URL", zap.String("url", baseURL), zap.Error(err))
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
			r.logger.Error("Failed to marshal cursor", zap.Error(err))
			return nil, errors.NewAppError("failed to marshal cursor", err)
		}
		queryParams.Set("cursor", string(cursorJSON))
	}
	parsedURL.RawQuery = queryParams.Encode()

	finalURL := parsedURL.String()
	r.logger.Debug("Making request to collection sync endpoint", zap.String("url", finalURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		r.logger.Error("Failed to create HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	r.logger.Debug("Executing HTTP request for collection sync")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to execute HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	r.logger.Debug("Parsing HTTP response body into CollectionSyncResponse")
	var response sync.CollectionSyncResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body into CollectionSyncResponse",
			zap.ByteString("body", body),
			zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("Successfully retrieved collection sync data from cloud",
		zap.Int("collections_count", len(response.Collections)),
		zap.Bool("has_more", response.HasMore))

	return &response, nil
}

func (r *syncRepository) GetFileSyncData(ctx context.Context, cursor *sync.SyncCursor, limit int64) (*sync.FileSyncResponse, error) {
	r.logger.Debug("Getting file sync data from cloud",
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/sync/files", serverURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error("Failed to parse base URL", zap.String("url", baseURL), zap.Error(err))
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
			r.logger.Error("Failed to marshal cursor", zap.Error(err))
			return nil, errors.NewAppError("failed to marshal cursor", err)
		}
		queryParams.Set("cursor", string(cursorJSON))
	}
	parsedURL.RawQuery = queryParams.Encode()

	finalURL := parsedURL.String()
	r.logger.Debug("Making request to file sync endpoint", zap.String("url", finalURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		r.logger.Error("Failed to create HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	r.logger.Debug("Executing HTTP request for file sync")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to execute HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	r.logger.Debug("Parsing HTTP response body into FileSyncResponse")
	var response sync.FileSyncResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body into FileSyncResponse",
			zap.ByteString("body", body),
			zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("Successfully retrieved file sync data from cloud",
		zap.Int("files_count", len(response.Files)),
		zap.Bool("has_more", response.HasMore))

	return &response, nil
}
