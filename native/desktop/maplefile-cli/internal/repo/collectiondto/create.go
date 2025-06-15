// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/upload.go
package collectiondto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

func (r *collectionDTORepository) CreateInCloud(ctx context.Context, collectionDTO *collectiondto.CollectionDTO) (*gocql.UUID, error) {
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	r.logger.Debug("🔍 Marshalling create collection request to JSON")
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address from config", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Convert request to JSON
	r.logger.Debug("✅ Successfully marshalled request to JSON")
	jsonData, err := json.Marshal(collectionDTO)
	if err != nil {
		r.logger.Error("❌ Failed to marshal request to JSON", zap.Any("request", collectionDTO), zap.Error(err))
		return nil, errors.NewAppError("failed to marshal request", err)
	}
	r.logger.Debug("✅ Successfully marshalled request to JSON")

	// Create HTTP request
	createURL := fmt.Sprintf("%s/maplefile/api/v1/collections", serverURL)
	r.logger.Info("➡️ Making HTTP request to create collection",
		zap.String("method", "POST"),
		zap.String("url", createURL))

	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request for creating collection", zap.String("url", createURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT "+accessToken) // Log this carefully if sensitive
	r.logger.Debug("🔍 HTTP request headers set")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request to create collection")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request to create collection", zap.String("url", createURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()
	r.logger.Info("⬅️ Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	r.logger.Debug("🔍 Reading HTTP response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read HTTP response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}
	r.logger.Debug("✅ Successfully read HTTP response body", zap.ByteString("body", body)) // Be careful logging raw body if sensitive

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		r.logger.Error("🚨 Server returned an error status code", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("🚨 Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}
	r.logger.Debug("✅ HTTP response status is successful")

	// Parse the response
	r.logger.Debug("🔍 Parsing HTTP response body into CollectionDTO")
	var response collectiondto.CollectionDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into CollectionDTO", zap.ByteString("body", body), zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}
	r.logger.Debug("✅ Successfully parsed HTTP response body")

	r.logger.Info("✨ Successfully created collection", zap.Any("response", response))
	return &response.ID, nil
}
