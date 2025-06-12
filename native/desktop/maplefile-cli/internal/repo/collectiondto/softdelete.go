// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/softdelete.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *collectionDTORepository) SoftDeleteInCloudByID(ctx context.Context, id gocql.UUID) error {
	r.logger.Debug("🗑️ (Soft)Deleting collection from cloud",
		zap.String("collectionID", id.String()))

	// Validate input
	if id.String() == "" {
		r.logger.Error("❌ Collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get access token
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get access token", zap.Error(err))
		return errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address", zap.Error(err))
		return errors.NewAppError("failed to get cloud provider address", err)
	}

	// Create delete URL
	deleteURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s", serverURL, id.String())
	r.logger.Info("➡️ Making HTTP request to delete collection",
		zap.String("method", "DELETE"),
		zap.String("url", deleteURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request for (soft)deleting collection",
			zap.String("url", deleteURL),
			zap.Error(err))
		return errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	r.logger.Debug("🔍 HTTP request headers set")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request to (soft)delete collection")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request to (soft)delete collection",
			zap.String("url", deleteURL),
			zap.Error(err))
		return errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Info("⬅️ Received HTTP response",
		zap.String("status", resp.Status),
		zap.Int("statusCode", resp.StatusCode))

	// Read response body
	r.logger.Debug("🔍 Reading HTTP response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read HTTP response body", zap.Error(err))
		return errors.NewAppError("failed to read response", err)
	}
	r.logger.Debug("✅ Successfully read HTTP response body", zap.ByteString("body", body))

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		// Success - collection deleted
		r.logger.Info("✅ Successfully (soft)deleted collection from cloud",
			zap.String("collectionID", id.String()))
		return nil

	case http.StatusNotFound:
		// Collection not found
		r.logger.Warn("⚠️ Collection not found on server",
			zap.String("collectionID", id.String()),
			zap.String("status", resp.Status))
		return errors.NewAppError("collection not found", nil)

	case http.StatusForbidden:
		// Permission denied
		r.logger.Error("🚫 Permission denied to (soft)delete collection",
			zap.String("collectionID", id.String()),
			zap.String("status", resp.Status))
		return errors.NewAppError("permission denied - you don't have rights to delete this collection", nil)

	case http.StatusUnauthorized:
		// Authentication failed
		r.logger.Error("🔐 Authentication failed",
			zap.String("collectionID", id.String()),
			zap.String("status", resp.Status))
		return errors.NewAppError("authentication failed - please login again", nil)

	default:
		// Other error status codes
		r.logger.Error("🚨 Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		// Try to parse error message from response body
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("🚨 Server returned error message in response body",
					zap.String("message", errMsg))
				return errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}

		return errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}
}
