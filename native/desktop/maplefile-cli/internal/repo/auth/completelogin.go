// monorepo/native/desktop/maplefile-cli/internal/repository/auth/completelogin.go
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

// completeLoginRepository implements CompleteLoginRepository interface
type completeLoginRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewCompleteLoginRepository creates a new repository for login completion
func NewCompleteLoginRepository(logger *zap.Logger, configService config.ConfigService) auth.CompleteLoginRepository {
	logger = logger.Named("CompleteLoginRepository")
	return &completeLoginRepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// CompleteLogin sends the login completion request to the server
func (r *completeLoginRepository) CompleteLogin(ctx context.Context, request *auth.CompleteLoginRequest) (*auth.TokenResponse, error) {
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
	completeURL := fmt.Sprintf("%s/iam/api/v1/complete-login", serverURL)
	r.logger.Debug("Making HTTP request", zap.String("url", completeURL))

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
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s\nResponse: %s", resp.Status, string(body)), nil)
	}

	// Parse the response
	var tokenResp auth.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, errors.NewAppError("error parsing token response", err)
	}

	return &tokenResp, nil
}
