// native/desktop/maplefile-cli/internal/repo/recoverydto/verify.go
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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// VerifyRecoveryFromCloud verifies the recovery challenge with the cloud service
func (r *recoveryDTORepository) VerifyRecoveryFromCloud(ctx context.Context, request *recovery.RecoveryVerifyRequest) (*recovery.RecoveryVerifyResponse, error) {
	r.logger.Debug("🔐 Verifying recovery challenge from cloud",
		zap.String("sessionID", request.SessionID))

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Validate request
	if err := recovery.ValidateRecoveryVerifyRequest(request); err != nil {
		r.logger.Error("❌ Invalid recovery verify request", zap.Error(err))
		return nil, err
	}

	// Convert request to JSON
	r.logger.Debug("🔍 Marshalling recovery verify request to JSON")
	jsonData, err := json.Marshal(request)
	if err != nil {
		r.logger.Error("❌ Failed to marshal request to JSON", zap.Any("request", request), zap.Error(err))
		return nil, errors.NewAppError("failed to marshal request", err)
	}
	r.logger.Debug("✅ Successfully marshalled request to JSON")

	// Create HTTP request
	verifyURL := fmt.Sprintf("%s/iam/api/v1/recovery/verify", serverURL)
	r.logger.Info("➡️ Making HTTP request to verify recovery",
		zap.String("method", "POST"),
		zap.String("url", verifyURL))

	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request for recovery verification", zap.String("url", verifyURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	r.logger.Debug("🔍 HTTP request headers set")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request to verify recovery")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request to verify recovery", zap.String("url", verifyURL), zap.Error(err))
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
			if details, ok := errorResponse["details"].(map[string]interface{}); ok {
				if challengeErr, ok := details["decrypted_challenge"].(string); ok {
					r.logger.Error("🚨 Challenge verification failed", zap.String("challengeError", challengeErr))
					return nil, recovery.NewChallengeDecryptionError()
				}
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}
	r.logger.Debug("✅ HTTP response status is successful")

	// Parse the response
	r.logger.Debug("🔍 Parsing HTTP response body into RecoveryVerifyResponse")
	var response recovery.RecoveryVerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse response body into RecoveryVerifyResponse", zap.ByteString("body", body), zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}
	r.logger.Debug("✅ Successfully parsed HTTP response body")

	r.logger.Info("✨ Successfully verified recovery challenge from cloud",
		zap.String("sessionID", request.SessionID),
		zap.Int("expiresIn", response.ExpiresIn))

	return &response, nil
}
