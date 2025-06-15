// native/desktop/maplefile-cli/internal/repo/auth/token.go
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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/authdto"
)

// tokenDTORepositoryImpl implements the TokenDTORepository interface
type tokenDTORepositoryImpl struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	userRepo               user.Repository
	tokenDecryptionService svc_authdto.TokenDecryptionService
}

// NewTokenDTORepository creates a new instance of TokenDTORepository
func NewTokenDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenDecryptionService svc_authdto.TokenDecryptionService,
) dom_authdto.TokenDTORepository {
	logger = logger.Named("TokenRepository")
	return &tokenDTORepositoryImpl{
		logger:                 logger,
		configService:          configService,
		userRepo:               userRepo,
		tokenDecryptionService: tokenDecryptionService,
	}
}

func (s *tokenDTORepositoryImpl) Save(
	ctx context.Context,
	email string,
	accessToken string,
	accessTokenExpiryDate *time.Time,
	refreshToken string,
	refreshTokenExpiryDate *time.Time,
) error {
	return s.configService.SetLoggedInUserCredentials(ctx, email, accessToken, accessTokenExpiryDate, refreshToken, refreshTokenExpiryDate)
}

func (s *tokenDTORepositoryImpl) GetAccessToken(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if creds.AccessToken == "" || time.Now().Add(30*time.Second).After(*creds.AccessTokenExpiryTime) {
		// Check if we have a refresh token
		if creds.RefreshToken == "" {
			return "", errors.NewAppError("no refresh token available", nil)
		}

		// Check if refresh token is still valid
		if time.Now().After(*creds.RefreshTokenExpiryTime) {
			return "", errors.NewAppError("refresh token has expired, please login again", nil)
		}

		// Refresh the token
		tokenResp, err := s.refreshFromCloud(ctx, creds.RefreshToken, creds.Email)
		if err != nil {
			return "", errors.NewAppError("failed to refresh token", err)
		}

		// Update credentials
		if err := s.configService.SetLoggedInUserCredentials(
			ctx,
			creds.Email,
			tokenResp.AccessToken,
			&tokenResp.AccessTokenExpiryDate,
			tokenResp.RefreshToken,
			&tokenResp.RefreshTokenExpiryDate,
		); err != nil {
			return "", errors.NewAppError("failed to save refreshed credentials", err)
		}

		creds.AccessToken = tokenResp.AccessToken
	}

	return creds.AccessToken, nil
}

func (s *tokenDTORepositoryImpl) GetAccessTokenAfterForcedRefresh(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Refresh the token
	tokenResp, err := s.refreshFromCloud(ctx, creds.RefreshToken, creds.Email)
	if err != nil {
		return "", errors.NewAppError("failed to refresh token", err)
	}

	// Update credentials
	if err := s.configService.SetLoggedInUserCredentials(
		ctx,
		creds.Email,
		tokenResp.AccessToken,
		&tokenResp.AccessTokenExpiryDate,
		tokenResp.RefreshToken,
		&tokenResp.RefreshTokenExpiryDate,
	); err != nil {
		return "", errors.NewAppError("failed to save refreshed credentials", err)
	}

	return tokenResp.AccessToken, nil
}

type TokenRefreshRequestDTO struct {
	Value string `json:"value"`
}

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

// RefreshToken refreshes the authentication token using the provided refresh token
func (s *tokenDTORepositoryImpl) refreshFromCloud(ctx context.Context, refreshToken string, email string) (*TokenRefreshResponseDTO, error) {
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

	// Check if we received encrypted tokens
	if tokenResponse.EncryptedTokens != "" {
		s.logger.Info("Received encrypted refresh tokens, need to decrypt",
			zap.String("email", email))

		// Get user data to decrypt tokens
		userData, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil || userData == nil {
			return nil, errors.NewAppError("failed to retrieve user data for token decryption", err)
		}

		// For token refresh, we need the user's password to decrypt
		// In a production system, you might want to:
		// 1. Store password temporarily in secure memory during session
		// 2. Use OS keychain/keyring services
		// 3. Prompt user for password
		// For now, we'll return an error indicating password is needed

		// Check if we have a password stored (this would be implemented with secure storage)
		password := s.getStoredPassword(ctx, email)
		if password == "" {
			// We can't decrypt without the password
			s.logger.Warn("Cannot decrypt refreshed tokens without password",
				zap.String("email", email))
			return nil, errors.NewAppError("password required to decrypt refreshed tokens - please login again", nil)
		}

		// Decrypt the tokens
		accessToken, refreshToken, err := s.tokenDecryptionService.DecryptTokens(
			tokenResponse.EncryptedTokens,
			userData,
			password,
		)
		if err != nil {
			return nil, errors.NewAppError("failed to decrypt refreshed tokens", err)
		}

		// Update response with decrypted tokens
		tokenResponse.AccessToken = accessToken
		tokenResponse.RefreshToken = refreshToken

		s.logger.Info("Successfully decrypted refreshed tokens",
			zap.String("email", email))
	} else if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		// No tokens received
		return nil, errors.NewAppError("no tokens received from refresh", nil)
	}

	return &tokenResponse, nil
}

// getStoredPassword retrieves stored password (placeholder - implement secure storage)
func (s *tokenDTORepositoryImpl) getStoredPassword(ctx context.Context, email string) string {
	// This is a placeholder - in production, implement secure password storage
	// Options:
	// 1. OS Keychain/Keyring integration
	// 2. Secure in-memory storage during session
	// 3. Hardware security module
	return ""
}
