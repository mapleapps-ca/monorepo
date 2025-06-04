// native/desktop/maplefile-cli/internal/repo/medto/get.go
package medto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
)

func (r *meDTORepository) GetMeFromCloud(ctx context.Context) (*medto.MeResponseDTO, error) {
	r.logger.Debug("🔍 Getting user profile from cloud")

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

	// Create HTTP request
	getURL := fmt.Sprintf("%s/maplefile/api/v1/me", serverURL)
	r.logger.Info("➡️ Making HTTP request to get user profile",
		zap.String("method", "GET"),
		zap.String("url", getURL))

	req, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request for getting user profile", zap.String("url", getURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	r.logger.Debug("🔍 HTTP request headers set")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request to get user profile")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request to get user profile", zap.String("url", getURL), zap.Error(err))
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
	r.logger.Debug("✅ Successfully read HTTP response body")

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
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
	r.logger.Debug("🔍 Parsing HTTP response body into MeResponseDTO")
	var response medto.MeResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into MeResponseDTO", zap.ByteString("body", body), zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}
	r.logger.Debug("✅ Successfully parsed HTTP response body")

	r.logger.Info("✨ Successfully retrieved user profile from cloud", zap.String("email", response.Email))
	return &response, nil
}
