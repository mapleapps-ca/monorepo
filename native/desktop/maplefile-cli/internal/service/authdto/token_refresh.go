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
	// GetValidAccessToken gets a valid access token, refreshing if needed
	GetValidAccessToken(ctx context.Context) (string, error)
	// RefreshTokenWithPassword refreshes and decrypts tokens using the provided password
	RefreshTokenWithPassword(ctx context.Context, password string) (string, error)
	// ForceRefresh forces a token refresh regardless of expiry
	ForceRefresh(ctx context.Context) (string, error)
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
	// Legacy plaintext fields
	Email                  string    `json:"username"`
	AccessToken            string    `json:"access_token,omitempty"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshToken           string    `json:"refresh_token,omitempty"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`

	// New encrypted token fields
	EncryptedTokens string `json:"encrypted_tokens,omitempty"`
	TokenNonce      string `json:"token_nonce,omitempty"`
}

// GetValidAccessToken gets a valid access token, refreshing if needed
func (s *tokenRefreshService) GetValidAccessToken(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if creds.AccessToken == "" || time.Now().Add(30*time.Second).After(*creds.AccessTokenExpiryTime) {
		s.logger.Info("Access token expired or expiring soon, attempting refresh",
			zap.String("email", creds.Email))

		// Try to refresh the token
		newToken, err := s.ForceRefresh(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		return newToken, nil
	}

	return creds.AccessToken, nil
}

// ForceRefresh forces a token refresh regardless of expiry
func (s *tokenRefreshService) ForceRefresh(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil || creds == nil {
		return "", errors.NewAppError("no logged in user found", err)
	}

	// Check if refresh token is still valid
	if time.Now().After(*creds.RefreshTokenExpiryTime) {
		return "", errors.NewAppError("refresh token has expired, please login again", nil)
	}

	s.logger.Info("Forcing token refresh", zap.String("email", creds.Email))

	// Call the cloud API to refresh tokens
	refreshResponse, err := s.refreshFromCloud(ctx, creds.RefreshToken, creds.Email)
	if err != nil {
		return "", fmt.Errorf("failed to refresh tokens from cloud: %w", err)
	}

	// Handle encrypted vs plaintext tokens
	var finalAccessToken, finalRefreshToken string

	if refreshResponse.EncryptedTokens != "" {
		s.logger.Info("Received encrypted tokens, decrypting...", zap.String("email", creds.Email))

		// We have encrypted tokens - need to decrypt them
		// Get user data for decryption
		userData, err := s.userRepo.GetByEmail(ctx, creds.Email)
		if err != nil || userData == nil {
			return "", errors.NewAppError("failed to retrieve user data for token decryption", err)
		}

		// For encrypted token refresh, we need the password
		// Since we don't have the password stored, we need to prompt for it
		// For now, return an error indicating password is needed
		return "", errors.NewAppError("encrypted tokens received - password required for decryption. Please login again or use RefreshTokenWithPassword", nil)

	} else if refreshResponse.AccessToken != "" && refreshResponse.RefreshToken != "" {
		// We have plaintext tokens (legacy mode)
		s.logger.Info("Received plaintext tokens", zap.String("email", creds.Email))
		finalAccessToken = refreshResponse.AccessToken
		finalRefreshToken = refreshResponse.RefreshToken

	} else {
		return "", errors.NewAppError("no valid tokens received from refresh", nil)
	}

	// Save the new tokens
	err = s.configService.SetLoggedInUserCredentials(
		ctx,
		creds.Email,
		finalAccessToken,
		&refreshResponse.AccessTokenExpiryDate,
		finalRefreshToken,
		&refreshResponse.RefreshTokenExpiryDate,
	)
	if err != nil {
		return "", errors.NewAppError("failed to save refreshed credentials", err)
	}

	s.logger.Info("Token refresh completed successfully", zap.String("email", creds.Email))
	return finalAccessToken, nil
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

	s.logger.Info("Refreshing tokens with password", zap.String("email", creds.Email))

	// Call the cloud API to refresh tokens
	refreshResponse, err := s.refreshFromCloud(ctx, creds.RefreshToken, creds.Email)
	if err != nil {
		return "", fmt.Errorf("failed to refresh tokens from cloud: %w", err)
	}

	var finalAccessToken, finalRefreshToken string

	if refreshResponse.EncryptedTokens != "" {
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

		finalAccessToken = accessToken
		finalRefreshToken = refreshToken

	} else if refreshResponse.AccessToken != "" && refreshResponse.RefreshToken != "" {
		// Plaintext tokens
		finalAccessToken = refreshResponse.AccessToken
		finalRefreshToken = refreshResponse.RefreshToken

	} else {
		return "", errors.NewAppError("no valid tokens received from refresh", nil)
	}

	// Save the new tokens
	err = s.configService.SetLoggedInUserCredentials(
		ctx,
		creds.Email,
		finalAccessToken,
		&refreshResponse.AccessTokenExpiryDate,
		finalRefreshToken,
		&refreshResponse.RefreshTokenExpiryDate,
	)
	if err != nil {
		return "", errors.NewAppError("failed to save refreshed credentials", err)
	}

	s.logger.Info("Token refresh with decryption completed successfully", zap.String("email", creds.Email))
	return finalAccessToken, nil
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

	return &tokenResponse, nil
}
