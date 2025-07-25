// cloud/mapleapps-backend/internal/iam/service/gateway/completelogin.go
package gateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_auth "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/auth"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/twotiercache"
)

// GatewayCompleteLoginRequestIDO used to get input from client containing the email, challenge ID, and the decrypted challenge
type GatewayCompleteLoginRequestIDO struct {
	Email         string `json:"email"`
	ChallengeID   string `json:"challengeId"`
	DecryptedData string `json:"decryptedData"`
}

// GatewayCompleteLoginResponseIDO is the response sent back to client with encrypted authentication tokens
type GatewayCompleteLoginResponseIDO struct {
	EncryptedAccessToken   string    `json:"encrypted_access_token"`
	EncryptedRefreshToken  string    `json:"encrypted_refresh_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
	TokenNonce             string    `json:"token_nonce"`
}

// Service interface for completing login
type GatewayCompleteLoginService interface {
	Execute(sessCtx context.Context, req *GatewayCompleteLoginRequestIDO) (*GatewayCompleteLoginResponseIDO, error)
}

// Implementation of complete login service
type gatewayCompleteLoginServiceImpl struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	cache                  twotiercache.TwoTierCacher
	jwtProvider            jwt.JWTProvider
	userGetByEmailUseCase  uc_user.FederatedUserGetByEmailUseCase
	userUpdateUseCase      uc_user.FederatedUserUpdateUseCase
	tokenEncryptionService dom_auth.TokenEncryptionService
}

func NewGatewayCompleteLoginService(
	config *config.Configuration,
	logger *zap.Logger,
	cache twotiercache.TwoTierCacher,
	jwtProvider jwt.JWTProvider,
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase,
	userUpdateUseCase uc_user.FederatedUserUpdateUseCase,
	tokenEncryptionService dom_auth.TokenEncryptionService,
) GatewayCompleteLoginService {
	logger = logger.Named("GatewayCompleteLoginService")
	return &gatewayCompleteLoginServiceImpl{
		config:                 config,
		logger:                 logger,
		cache:                  cache,
		jwtProvider:            jwtProvider,
		userGetByEmailUseCase:  userGetByEmailUseCase,
		userUpdateUseCase:      userUpdateUseCase,
		tokenEncryptionService: tokenEncryptionService,
	}
}

func (s *gatewayCompleteLoginServiceImpl) Execute(sessCtx context.Context, req *GatewayCompleteLoginRequestIDO) (*GatewayCompleteLoginResponseIDO, error) {
	// Validate input
	e := make(map[string]string)
	if req.Email == "" {
		e["email"] = "Email address is required"
	}
	if req.ChallengeID == "" {
		e["challengeId"] = "Challenge ID is required"
	}
	if req.DecryptedData == "" {
		e["decryptedData"] = "Decrypted data is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Sanitize input
	req.Email = strings.ToLower(req.Email)
	req.Email = strings.ReplaceAll(req.Email, " ", "")

	// Retrieve challenge data from cache
	challengeCacheKey := fmt.Sprintf("login_challenge:%s", req.ChallengeID)
	challengeDataJSON, err := s.cache.Get(sessCtx, challengeCacheKey)
	if err != nil {
		s.logger.Error("Failed to retrieve challenge data", zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("challengeId", "Invalid or expired challenge")
	}

	if challengeDataJSON == nil {
		s.logger.Error("Challenge data not found in cache")
		return nil, httperror.NewForBadRequestWithSingleField("challengeId", "Invalid or expired challenge")
	}

	// Unmarshal the data from JSON
	var challengeData ChallengeData
	if err := json.Unmarshal(challengeDataJSON, &challengeData); err != nil {
		s.logger.Error("Failed to unmarshal challenge data", zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("challengeId", "Invalid challenge")
	}

	// Verify the challenge
	if challengeData.Email != req.Email {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not match challenge")
	}

	// Check expiry
	if time.Now().After(challengeData.ExpiresAt) {
		return nil, httperror.NewForBadRequestWithSingleField("challengeId", "Challenge has expired")
	}

	// Check if already verified
	if challengeData.IsVerified {
		return nil, httperror.NewForBadRequestWithSingleField("challengeId", "Challenge has already been used")
	}

	// Verify the decrypted data by comparing the raw bytes of the challenge
	storedChallengeBase64 := challengeData.Challenge
	receivedChallengeBase64 := req.DecryptedData

	s.logger.Info("CompleteLogin: Verifying challenges",
		zap.String("challenge_id", req.ChallengeID),
		zap.String("stored_challenge_std_base64_from_cache", storedChallengeBase64),
		zap.String("received_challenge_urlsafe_base64_from_client", receivedChallengeBase64),
	)

	// Decode the stored challenge (standard Base64)
	storedChallengeBytes, err := base64.StdEncoding.DecodeString(storedChallengeBase64)
	if err != nil {
		// Try with RawURLEncoding as fallback
		storedChallengeBytes, err = base64.RawURLEncoding.DecodeString(storedChallengeBase64)
		if err != nil {
			s.logger.Error("Failed to decode stored challenge after trying multiple encodings",
				zap.String("challenge_id", challengeData.ChallengeID),
				zap.String("base64_value", storedChallengeBase64),
				zap.Error(err))
			return nil, httperror.NewForInternalServerError("Failed to process challenge due to internal data error")
		}
	}

	// Try both standard and URL-safe decoding if needed
	receivedChallengeBytes, err := base64.RawURLEncoding.DecodeString(receivedChallengeBase64)
	if err != nil {
		// Try with StdEncoding as fallback
		receivedChallengeBytes, err = base64.StdEncoding.DecodeString(receivedChallengeBase64)
		if err != nil {
			s.logger.Warn("Failed to decode received challenge after trying multiple encodings",
				zap.String("challenge_id", req.ChallengeID),
				zap.String("base64_value", receivedChallengeBase64),
				zap.Error(err))
			return nil, httperror.NewForBadRequestWithSingleField("decryptedData", "Invalid format for decrypted challenge")
		}
	}

	// Compare the raw byte slices
	if !bytes.Equal(storedChallengeBytes, receivedChallengeBytes) {
		s.logger.Error("Challenge verification failed: byte content mismatch after decoding",
			zap.String("challenge_id", req.ChallengeID),
		)
		return nil, httperror.NewForBadRequestWithSingleField("decryptedData", "Invalid challenge response")
	}

	s.logger.Info("CompleteLogin: Challenge verified successfully by byte comparison", zap.String("challenge_id", req.ChallengeID))

	// Get user from database
	user, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exist")
	}

	// Update last login timestamp if needed
	user.ModifiedAt = time.Now()
	if err := s.userUpdateUseCase.Execute(sessCtx, user); err != nil {
		s.logger.Warn("Failed to update user last login time", zap.Error(err))
		// Continue anyway, as this is not critical
	}

	// Mark challenge as verified
	challengeData.IsVerified = true

	// Marshal the updated challenge data to JSON
	updatedChallengeDataJSON, err := json.Marshal(challengeData)
	if err != nil {
		s.logger.Error("Failed to marshal updated challenge data", zap.Error(err))
		// Continue anyway, as this is not critical
	} else {
		if err := s.cache.SetWithExpiry(sessCtx, challengeCacheKey, updatedChallengeDataJSON, 5*time.Minute); err != nil {
			s.logger.Warn("Failed to update challenge in cache", zap.Error(err))
			// Continue anyway, as this is not critical
		}
	}

	// Generate JWT tokens and encrypt them
	return s.generateEncryptedTokens(sessCtx, user)
}

// generateEncryptedTokens creates access and refresh tokens and encrypts them separately with user's public key
func (s *gatewayCompleteLoginServiceImpl) generateEncryptedTokens(ctx context.Context, user *domain.FederatedUser) (*GatewayCompleteLoginResponseIDO, error) {
	// Convert user to JSON for storage in cache
	userBin, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Set expiry durations
	atExpiry := 30 * time.Minute    // 30 minutes
	rtExpiry := 14 * 24 * time.Hour // 14 days

	// Create a unique session ID
	sessionUUID := gocql.TimeUUID()

	// Store user data in cache
	err = s.cache.SetWithExpiry(ctx, sessionUUID.String(), userBin, rtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Generate JWT tokens
	accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry, err := s.jwtProvider.GenerateJWTTokenPair(sessionUUID.String(), atExpiry, rtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Require user to have a public key for encryption
	if user.SecurityData == nil || len(user.SecurityData.PublicKey.Key) == 0 {
		s.logger.Error("User does not have public key required for encryption",
			zap.String("email", user.Email))
		return nil, fmt.Errorf("user account not properly configured for secure authentication")
	}

	// Encrypt tokens separately with user's public key
	encryptedResponse, err := s.tokenEncryptionService.EncryptTokens(
		accessToken,
		refreshToken,
		user.SecurityData.PublicKey.Key,
		accessTokenExpiry,
		refreshTokenExpiry,
	)
	if err != nil {
		s.logger.Error("Failed to encrypt tokens",
			zap.Error(err),
			zap.String("email", user.Email))
		return nil, fmt.Errorf("failed to encrypt authentication tokens: %w", err)
	}

	// Return separately encrypted tokens
	return &GatewayCompleteLoginResponseIDO{
		EncryptedAccessToken:   encryptedResponse.EncryptedAccessToken,
		EncryptedRefreshToken:  encryptedResponse.EncryptedRefreshToken,
		TokenNonce:             encryptedResponse.Nonce,
		AccessTokenExpiryDate:  accessTokenExpiry,
		RefreshTokenExpiryDate: refreshTokenExpiry,
	}, nil
}
