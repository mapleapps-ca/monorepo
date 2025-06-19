// native/desktop/maplefile-cli/internal/service/authdto/token_refresh.go
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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// TokenRefreshService handles token refresh with encryption support
type TokenRefreshService interface {
	// RefreshTokenWithPassword refreshes and decrypts tokens using the provided password
	RefreshTokenWithPassword(ctx context.Context, password string) (string, error)
	// GetValidAccessToken gets a valid access token, refreshing if needed (requires password for encrypted tokens)
	GetValidAccessToken(ctx context.Context, password string) (string, error)
}

// tokenRefreshService implements TokenRefreshService
type tokenRefreshService struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	userRepo               user.Repository
	tokenDecryptionService TokenDecryptionService
}

// NewTokenRefreshService creates a new token refresh service
func NewTokenRefreshService(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenDecryptionService TokenDecryptionService,
) TokenRefreshService {
	logger = logger.Named("TokenRefreshService")
	return &tokenRefreshService{
		logger:                 logger,
		configService:          configService,
		userRepo:               userRepo,
		tokenDecryptionService: tokenDecryptionService,
	}
}

// TokenRefreshRequestDTO represents the request for token refresh
type TokenRefreshRequestDTO struct {
	Value string `json:"value"`
}

// TokenRefreshResponseDTO represents the response from token refresh
type TokenRefreshResponseDTO struct {
	Email                  string    `json:"username"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
	EncryptedTokens        string    `json:"encrypted_tokens"`
	TokenNonce             string    `json:"token_nonce"`
}

// GetValidAccessToken gets a valid access token, refreshing if needed
func (s *tokenRefreshService) GetValidAccessToken(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", errors.NewAppError("password is required for encrypted token operations", nil)
	}

	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if creds.AccessToken == "" || time.Now().Add(30*time.Second).After(*creds.AccessTokenExpiryTime) {
		s.logger.Info("Access token expired or expiring soon, refreshing",
			zap.String("email", creds.Email))

		// Refresh the token
		newToken, err := s.RefreshTokenWithPassword(ctx, password)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		return newToken, nil
	}

	return creds.AccessToken, nil
}

// RefreshTokenWithPassword refreshes and decrypts tokens using the provided password
func (s *tokenRefreshService) RefreshTokenWithPassword(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", errors.NewAppError("password is required for encrypted token refresh", nil)
	}

	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil || creds == nil {
		return "", errors.NewAppError("no logged in user found", err)
	}

	// Check if refresh token is still valid
	if time.Now().After(*creds.RefreshTokenExpiryTime) {
		return "", errors.NewAppError("refresh token has expired, please login again", nil)
	}

	s.logger.Info("Refreshing encrypted tokens", zap.String("email", creds.Email))

	// Call the cloud API to refresh tokens
	refreshResponse, err := s.refreshFromCloud(ctx, creds.RefreshToken, creds.Email)
	if err != nil {
		return "", fmt.Errorf("failed to refresh tokens from cloud: %w", err)
	}

	// Verify we received encrypted tokens
	if refreshResponse.EncryptedTokens == "" {
		return "", errors.NewAppError("server did not return encrypted tokens - this should not happen in encrypted-only mode", nil)
	}

	s.logger.Info("Decrypting received encrypted tokens", zap.String("email", creds.Email))

	// Get user data for decryption
	userData, err := s.userRepo.GetByEmail(ctx, creds.Email)
	if err != nil || userData == nil {
		return "", errors.NewAppError("failed to retrieve user data for token decryption", err)
	}

	// Decrypt the tokens using the user's private key
	accessToken, refreshToken, err := s.tokenDecryptionService.DecryptTokens(
		refreshResponse.EncryptedTokens,
		userData,
		password,
	)
	if err != nil {
		return "", errors.NewAppError("failed to decrypt refreshed tokens", err)
	}

	// Save the new tokens
	err = s.configService.SetLoggedInUserCredentials(
		ctx,
		creds.Email,
		accessToken,
		&refreshResponse.AccessTokenExpiryDate,
		refreshToken,
		&refreshResponse.RefreshTokenExpiryDate,
	)
	if err != nil {
		return "", errors.NewAppError("failed to save refreshed credentials", err)
	}

	s.logger.Info("Token refresh with decryption completed successfully", zap.String("email", creds.Email))
	return accessToken, nil
}

// refreshFromCloud calls the cloud API to refresh tokens
func (s *tokenRefreshService) refreshFromCloud(ctx context.Context, refreshToken string, email string) (*TokenRefreshResponseDTO, error) {
	// Get the server URL from configuration
	serverURL, err := s.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading cloud provider address: %w", err)
	}

	// Create the request payload
	refreshReq := TokenRefreshRequestDTO{
		Value: refreshToken,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(refreshReq)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Make HTTP request to server
	refreshURL := fmt.Sprintf("%s/iam/api/v1/token/refresh", serverURL)
	s.logger.Debug("Making token refresh request", zap.String("url", refreshURL))

	// Create and execute the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", refreshURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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
	var tokenResponse TokenRefreshResponseDTO
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Validate that we received encrypted tokens
	if tokenResponse.EncryptedTokens == "" {
		return nil, fmt.Errorf("server did not return encrypted tokens")
	}

	return &tokenResponse, nil
}
