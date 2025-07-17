// monorepo/cloud/mapleapps-backend/internal/iam/service/gateway/refreshtok.go
package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_auth "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/auth"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
	"go.uber.org/zap"
)

type GatewayRefreshTokenService interface {
	Execute(
		sessCtx context.Context,
		req *GatewayRefreshTokenRequestIDO,
	) (*GatewayRefreshTokenResponseIDO, error)
}

type gatewayRefreshTokenServiceImpl struct {
	logger                 *zap.Logger
	cache                  cassandracache.CassandraCacher
	jwtProvider            jwt.JWTProvider
	userGetByEmailUseCase  uc_user.FederatedUserGetByEmailUseCase
	tokenEncryptionService dom_auth.TokenEncryptionService
}

func NewGatewayRefreshTokenService(
	logger *zap.Logger,
	cach cassandracache.CassandraCacher,
	jwtp jwt.JWTProvider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
	tokenEncryptionService dom_auth.TokenEncryptionService,
) GatewayRefreshTokenService {
	logger = logger.Named("GatewayRefreshTokenService")
	return &gatewayRefreshTokenServiceImpl{logger, cach, jwtp, uc1, tokenEncryptionService}
}

type GatewayRefreshTokenRequestIDO struct {
	Value string `json:"value"`
}

// GatewayRefreshTokenResponseIDO struct used to represent the system's response when the refresh token request was a success.
type GatewayRefreshTokenResponseIDO struct {
	Email                  string    `json:"username"`
	EncryptedAccessToken   string    `json:"encrypted_access_token"`
	EncryptedRefreshToken  string    `json:"encrypted_refresh_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
	TokenNonce             string    `json:"token_nonce"`
}

func (s *gatewayRefreshTokenServiceImpl) Execute(
	sessCtx context.Context,
	req *GatewayRefreshTokenRequestIDO,
) (*GatewayRefreshTokenResponseIDO, error) {
	////
	//// Extract the `sessionID` so we can process it.
	////

	sessionID, err := s.jwtProvider.ProcessJWTToken(req.Value)
	if err != nil {
		err := errors.New("jwt refresh token failed")
		return nil, err
	}

	////
	//// Lookup in our in-memory the federateduser record for the `sessionID` or error.
	////

	uBin, err := s.cache.Get(sessCtx, sessionID)
	if err != nil {
		return nil, err
	}

	var u *domain.FederatedUser
	err = json.Unmarshal(uBin, &u)
	if err != nil {
		return nil, err
	}

	////
	//// Generate new access and refresh tokens and return them.
	////

	// Set expiry duration.
	atExpiry := 30 * time.Minute    // 30 minutes
	rtExpiry := 14 * 24 * time.Hour // 14 days

	// Start our session using an access and refresh token.
	newSessionUUID := gocql.TimeUUID().String()

	err = s.cache.SetWithExpiry(sessCtx, newSessionUUID, uBin, rtExpiry)
	if err != nil {
		return nil, err
	}

	// Generate our JWT token.
	accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry, err := s.jwtProvider.GenerateJWTTokenPair(newSessionUUID, atExpiry, rtExpiry)
	if err != nil {
		return nil, err
	}

	// Require user to have a public key for encryption
	if u.SecurityData == nil || len(u.SecurityData.PublicKey.Key) == 0 {
		s.logger.Error("User does not have public key required for encryption",
			zap.String("email", u.Email))
		return nil, errors.New("user account not properly configured for secure authentication")
	}

	// Encrypt tokens separately with user's public key
	encryptedResponse, err := s.tokenEncryptionService.EncryptTokens(
		accessToken,
		refreshToken,
		u.SecurityData.PublicKey.Key,
		accessTokenExpiry,
		refreshTokenExpiry,
	)
	if err != nil {
		s.logger.Error("Failed to encrypt tokens",
			zap.Error(err),
			zap.String("email", u.Email))
		return nil, fmt.Errorf("failed to encrypt authentication tokens: %w", err)
	}

	// Return separately encrypted tokens
	return &GatewayRefreshTokenResponseIDO{
		Email:                  u.Email,
		EncryptedAccessToken:   encryptedResponse.EncryptedAccessToken,
		EncryptedRefreshToken:  encryptedResponse.EncryptedRefreshToken,
		TokenNonce:             encryptedResponse.Nonce,
		AccessTokenExpiryDate:  accessTokenExpiry,
		RefreshTokenExpiryDate: refreshTokenExpiry,
	}, nil
}
