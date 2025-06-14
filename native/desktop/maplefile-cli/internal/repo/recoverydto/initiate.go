// native/desktop/maplefile-cli/internal/repo/recoverydto/initiate.go
package recoverydto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recoverydto"
)

// InitiateRecoveryFromCloud initiates account recovery with the cloud service
func (r *recoveryDTORepository) InitiateRecoveryFromCloud(ctx context.Context, request *recoverydto.RecoveryInitiateRequestDTO) (*recoverydto.RecoveryInitiateResponseDTO, error) {
	r.logger.Debug("🔐 Initiating account recovery from cloud",
		zap.String("email", request.Email),
		zap.String("method", request.Method))

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Validate request
	if err := recoverydto.ValidateRecoveryInitiateRequestDTO(request); err != nil {
		r.logger.Error("❌ Invalid recovery initiate request", zap.Error(err))
		return nil, err
	}

	// Convert request to JSON
	r.logger.Debug("🔍 Marshalling recovery initiate request to JSON")
	jsonData, err := json.Marshal(request)
	if err != nil {
		r.logger.Error("❌ Failed to marshal request to JSON", zap.Any("request", request), zap.Error(err))
		return nil, errors.NewAppError("failed to marshal request", err)
	}
	r.logger.Debug("✅ Successfully marshalled request to JSON")

	// Create HTTP request
	initiateURL := fmt.Sprintf("%s/iam/api/v1/recovery/initiate", serverURL)
	r.logger.Info("➡️ Making HTTP request to initiate recovery",
		zap.String("method", "POST"),
		zap.String("url", initiateURL))

	req, err := http.NewRequestWithContext(ctx, "POST", initiateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request for recovery initiation", zap.String("url", initiateURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	r.logger.Debug("🔍 HTTP request headers set")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request to initiate recovery")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request to initiate recovery", zap.String("url", initiateURL), zap.Error(err))
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
	r.logger.Debug("✅ HTTP response status is successful")

	// Parse the response
	r.logger.Debug("🔍 Parsing HTTP response body into RecoveryInitiateResponseDTO")
	var response recoverydto.RecoveryInitiateResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into RecoveryInitiateResponseDTO", zap.ByteString("body", body), zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}
	r.logger.Debug("✅ Successfully parsed HTTP response body")

	r.logger.Info("✨ Successfully initiated account recovery from cloud",
		zap.String("sessionID", response.SessionID),
		zap.String("challengeID", response.ChallengeID),
		zap.Int("expiresIn", response.ExpiresIn))

	return &response, nil
}
