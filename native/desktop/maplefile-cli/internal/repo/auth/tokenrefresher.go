// internal/repo/auth/tokenrefresher.go
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

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
)

// tokenRefresherRepo implements the auth.TokenRefresher interface
type tokenRefresherRepo struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// TokenRefreshRequest represents the data structure sent to the token refresh endpoint
type TokenRefreshRequest struct {
	Value string `json:"value"`
}

// NewTokenRefresherRepo creates a new instance of a TokenRefresher
func NewTokenRefresherRepo(logger *zap.Logger, configService config.ConfigService) auth.TokenRefresher {
	return &tokenRefresherRepo{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// RefreshToken refreshes the authentication token using the provided refresh token
func (r *tokenRefresherRepo) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenRefreshResponse, error) {
	// Get the server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading cloud provider address: %w", err)
	}

	// Create the request payload
	refreshReq := TokenRefreshRequest{
		Value: refreshToken,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(refreshReq)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Make HTTP request to server
	refreshURL := fmt.Sprintf("%s/iam/api/v1/token/refresh", serverURL)
	r.logger.Debug("Connecting to refresh token endpoint", zap.String("url", refreshURL))

	// Create and execute the HTTP request
	req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check response status code
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Try to parse error message if available
		var errorResponse map[string]any
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, fmt.Errorf("server error: %s", errMsg)
			}
		}
		return nil, fmt.Errorf("server returned error status: %s", resp.Status)
	}

	// Parse the response
	var respData struct {
		Email                  string    `json:"username"`
		AccessToken            string    `json:"access_token"`
		AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
		RefreshToken           string    `json:"refresh_token"`
		RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &auth.TokenRefreshResponse{
		AccessToken:            respData.AccessToken,
		AccessTokenExpiryDate:  respData.AccessTokenExpiryDate,
		RefreshToken:           respData.RefreshToken,
		RefreshTokenExpiryDate: respData.RefreshTokenExpiryDate,
	}, nil
}
