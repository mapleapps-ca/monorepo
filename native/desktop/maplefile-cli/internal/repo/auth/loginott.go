// monorepo/native/desktop/maplefile-cli/internal/repo/auth/loginott.go
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

// loginOTTRepository implements LoginOTTRepository interface
type loginOTTRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewLoginOTTRepository creates a new repository for login OTT operations
func NewLoginOTTRepository(logger *zap.Logger, configService config.ConfigService) auth.LoginOTTRepository {
	return &loginOTTRepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// RequestLoginOTT makes the HTTP request to get a login OTT
func (r *loginOTTRepository) RequestLoginOTT(ctx context.Context, request *auth.LoginOTTRequest) (*auth.LoginOTTResponse, error) {
	// Get server URL from config
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create the HTTP request
	requestURL := fmt.Sprintf("%s/iam/api/v1/request-ott", serverURL)
	r.logger.Debug("Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
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

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned status: %s", resp.Status), nil)
	}

	// Success response
	return &auth.LoginOTTResponse{
		Success: true,
		Message: "One-time login token request successful",
	}, nil
}
