// monorepo/native/desktop/maplefile-cli/internal/repo/auth/recovery.go
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
)

// recoveryRepository implements RecoveryRepository interface
type recoveryRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewRecoveryRepository creates a new repository for recovery operations
func NewRecoveryRepository(logger *zap.Logger, configService config.ConfigService) auth.RecoveryRepository {
	logger = logger.Named("RecoveryRepository")
	return &recoveryRepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// InitiateRecovery starts the recovery process
func (r *recoveryRepository) InitiateRecovery(ctx context.Context, request *auth.RecoveryRequest) (*auth.RecoveryVerifyResponse, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("error creating request", err)
	}

	// Create HTTP request
	initiateURL := fmt.Sprintf("%s/iam/api/v1/recovery/initiate", serverURL)
	r.logger.Debug("🔐 Initiating account recovery", zap.String("url", initiateURL))

	req, err := http.NewRequestWithContext(ctx, "POST", initiateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("error creating HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("error connecting to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("error reading response", err)
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
	var verifyResp auth.RecoveryVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, errors.NewAppError("error parsing recovery response", err)
	}

	r.logger.Info("✅ Recovery initiation successful", zap.String("sessionID", verifyResp.SessionID))
	return &verifyResp, nil
}

// CompleteRecovery completes the recovery process with new password
func (r *recoveryRepository) CompleteRecovery(ctx context.Context, request *auth.RecoveryCompleteRequest) (*auth.RecoveryCompleteResponse, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("error creating request", err)
	}

	// Create HTTP request
	completeURL := fmt.Sprintf("%s/iam/api/v1/recovery/complete", serverURL)
	r.logger.Debug("🔐 Completing account recovery", zap.String("url", completeURL))

	req, err := http.NewRequestWithContext(ctx, "POST", completeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("error creating HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("error connecting to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("error reading response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s\nResponse: %s", resp.Status, string(body)), nil)
	}

	// Parse the response
	var completeResp auth.RecoveryCompleteResponse
	if err := json.Unmarshal(body, &completeResp); err != nil {
		return nil, errors.NewAppError("error parsing recovery complete response", err)
	}

	return &completeResp, nil
}
