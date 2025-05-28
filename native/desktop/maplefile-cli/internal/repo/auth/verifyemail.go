// monorepo/native/desktop/maplefile-cli/internal/repo/auth/verifyemail.go
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

// emailVerificationRepository implements EmailVerificationRepository interface
type emailVerificationRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewEmailVerificationRepository creates a new repository for email verification
func NewEmailVerificationRepository(logger *zap.Logger, configService config.ConfigService) auth.EmailVerificationRepository {
	logger = logger.Named("EmailVerificationRepository")
	return &emailVerificationRepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// VerifyEmail sends a verification request to the server
func (r *emailVerificationRepository) VerifyEmail(ctx context.Context, code string) (*auth.VerifyEmailResponse, error) {
	// Get server URL from config
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Create the request payload
	verifyReq := auth.VerifyEmailRequest{
		Code: code,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(verifyReq)
	if err != nil {
		return nil, errors.NewAppError("failed to create request", err)
	}

	// Construct the URL
	verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-email-code", serverURL)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			} else if errField, ok := errorResponse["code"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errField), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response auth.VerifyEmailResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	return &response, nil
}
