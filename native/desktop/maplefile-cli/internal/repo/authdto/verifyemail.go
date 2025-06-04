// monorepo/native/desktop/maplefile-cli/internal/repo/auth/verifyemail.go
package authdto

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
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
)

// emailVerificationDTORepository implements EmailVerificationRepository interface
type emailVerificationDTORepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewEmailVerificationDTORepository creates a new repository for email verification
func NewEmailVerificationDTORepository(logger *zap.Logger, configService config.ConfigService) dom_authdto.EmailVerificationDTORepository {
	logger = logger.Named("EmailVerificationDTORepository")
	return &emailVerificationDTORepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// VerifyEmail sends a verification request to the server
func (r *emailVerificationDTORepository) VerifyEmail(ctx context.Context, code string) (*dom_authdto.VerifyEmailResponseDTO, error) {
	r.logger.Debug("✨ Starting email verification process")

	// Get server URL from config
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address for verification", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}
	r.logger.Debug("➡️ Retrieved server URL", zap.String("serverURL", serverURL))

	// Create the request payload
	verifyReq := dom_authdto.VerifyEmailRequestDTO{
		Code: code,
	}
	r.logger.Debug("➡️ Created verification request payload")

	// Convert request to JSON
	jsonData, err := json.Marshal(verifyReq)
	if err != nil {
		r.logger.Error("❌ Failed to marshal verification request", zap.Error(err))
		return nil, errors.NewAppError("failed to create request", err)
	}
	r.logger.Debug("➡️ Marshalled request payload to JSON")

	// Construct the URL
	verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-email-code", serverURL)
	r.logger.Debug("➡️ Constructed verification URL", zap.String("verifyURL", verifyURL))

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	r.logger.Debug("➡️ Created HTTP POST request with JSON header")

	// Execute the request
	r.logger.Debug("➡️ Sending HTTP request to server")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to connect to server for verification", zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()
	r.logger.Debug("➡️ Received HTTP response", zap.Int("statusCode", resp.StatusCode))

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}
	r.logger.Debug("➡️ Read response body")

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		r.logger.Warn("⚠️ Server returned non-success status code", zap.Int("statusCode", resp.StatusCode), zap.ByteString("responseBody", body))
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("❌ Server error message", zap.String("errorMessage", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			} else if errField, ok := errorResponse["code"].(string); ok {
				r.logger.Error("❌ Server error code", zap.String("errorCode", errField))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errField), nil)
			}
		}
		r.logger.Error("❌ Server returned error status without recognizable error body", zap.String("status", resp.Status))
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}
	r.logger.Debug("➡️ Server returned success status code")

	// Parse the response
	var response dom_authdto.VerifyEmailResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse success response body", zap.Error(err), zap.ByteString("responseBody", body))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Email verification successful")
	return &response, nil
}
