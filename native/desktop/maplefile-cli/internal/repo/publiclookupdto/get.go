// monorepo/native/desktop/maplefile-cli/internal/repo/publiclookupdto/get.go
package publiclookupdto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
)

func (r *publicLookupDTORepository) GetFromCloud(ctx context.Context, req *publiclookupdto.PublicLookupRequestDTO) (*publiclookupdto.PublicLookupResponseDTO, error) {
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("🚨 Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("🚨 Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Defensive programming
	if req.Email == "" {
		r.logger.Error("🚨 email is required")
		return nil, errors.NewAppError("email is required", nil)
	}

	// Create HTTP request
	publicUserLookupURL := fmt.Sprintf("%s/iam/api/v1/users/lookup?email=%s", serverURL, req.Email)
	request, err := http.NewRequestWithContext(ctx, "GET", publicUserLookupURL, nil)
	if err != nil {
		r.logger.Error("🚨 Failed to create HTTP request",
			zap.String("url", publicUserLookupURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	request.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(request)
	if err != nil {
		r.logger.Error("🚨 Failed to execute HTTP request", zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("🚨 Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(body), "email") {
			r.logger.Warn("⚠️ Server returned email not found error")
			return nil, errors.NewAppError("email does not exist", nil)
		}
		r.logger.Error("🚨 Server returned an error status code",
			zap.String("status", resp.Status),
			zap.String("body", string(body)),
			zap.Int("statusCode", resp.StatusCode))
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s\nResponse: %s", resp.Status, string(body)), nil)
	}

	// Parse the response
	var response publiclookupdto.PublicLookupResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("🚨 Failed to parse response body", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✨ Successfully fetched collection from cloud server",
		zap.String("email", req.Email))
	return &response, nil
}
