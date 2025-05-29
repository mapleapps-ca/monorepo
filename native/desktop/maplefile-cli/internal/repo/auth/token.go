// native/desktop/maplefile-cli/internal/repo/auth/token.go
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

// tokenRepositoryImpl implements the TokenRepository interface
type tokenRepositoryImpl struct {
	logger        *zap.Logger
	configService config.ConfigService
}

// NewTokenRepository creates a new instance of TokenRepository
func NewTokenRepository(logger *zap.Logger, configService config.ConfigService) auth.TokenRepository {
	logger = logger.Named("TokenRepository")
	return &tokenRepositoryImpl{
		logger:        logger,
		configService: configService,
	}
}

func (s *tokenRepositoryImpl) Save(
	ctx context.Context,
	email string,
	accessToken string,
	accessTokenExpiryDate *time.Time,
	refreshToken string,
	refreshTokenExpiryDate *time.Time,
) error {
	return s.configService.SetLoggedInUserCredentials(ctx, email, accessToken, accessTokenExpiryDate, refreshToken, refreshTokenExpiryDate)
}

func (s *tokenRepositoryImpl) GetAccessToken(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if creds.AccessToken == "" || time.Now().Add(30*time.Second).After(*creds.AccessTokenExpiryTime) {
		// r.logger.Info("Access token expired or will expire soon, attempting to refresh",
		// 	zap.String("email", creds.Email),
		// 	zap.Time("tokenExpiry", creds.AccessTokenExpiryTime))

		// Check if we have a refresh token
		if creds.RefreshToken == "" {
			return "", errors.NewAppError("no refresh token available", nil)
		}

		// Check if refresh token is still valid
		if time.Now().After(*creds.RefreshTokenExpiryTime) {
			return "", errors.NewAppError("refresh token has expired, please login again", nil)
		}

		// Refresh the token using the domain interface
		tokenResp, err := s.refreshFromCloud(ctx, creds.RefreshToken)
		if err != nil {
			return "", errors.NewAppError("failed to refresh token", err)
		}
		if err := s.configService.SetLoggedInUserCredentials(ctx, creds.Email, tokenResp.AccessToken, &tokenResp.AccessTokenExpiryDate, tokenResp.RefreshToken, &tokenResp.RefreshTokenExpiryDate); err != nil {
			return "", errors.NewAppError("failed to save refreshed credentials", err)
		}
		creds.AccessToken = tokenResp.AccessToken
	}

	return creds.AccessToken, nil
}

func (s *tokenRepositoryImpl) GetAccessTokenAfterForcedRefresh(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Refresh the token using the domain interface
	tokenResp, err := s.refreshFromCloud(ctx, creds.RefreshToken)
	if err != nil {
		return "", errors.NewAppError("failed to refresh token", err)
	}
	if err := s.configService.SetLoggedInUserCredentials(ctx, creds.Email, tokenResp.AccessToken, &tokenResp.AccessTokenExpiryDate, tokenResp.RefreshToken, &tokenResp.RefreshTokenExpiryDate); err != nil {
		return "", errors.NewAppError("failed to save refreshed credentials", err)
	}
	creds.AccessToken = tokenResp.AccessToken
	return creds.AccessToken, nil
}

type TokenRefreshRequestDTO struct {
	Value string `json:"value"`
}

type TokenRefreshResponseDTO struct {
	Email                  string    `json:"username"`
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
}

// RefreshToken refreshes the authentication token using the provided refresh token
func (s *tokenRepositoryImpl) refreshFromCloud(ctx context.Context, refreshToken string) (*TokenRefreshResponseDTO, error) {
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
	s.logger.Debug("🌐 Connecting to refresh token endpoint", zap.String("url", refreshURL))

	// Create and execute the HTTP request
	req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer(jsonData))
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
