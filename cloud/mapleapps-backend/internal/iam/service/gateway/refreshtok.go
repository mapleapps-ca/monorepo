package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gocql/gocql"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type GatewayRefreshTokenService interface {
	Execute(
		sessCtx context.Context,
		req *GatewayRefreshTokenRequestIDO,
	) (*GatewayRefreshTokenResponseIDO, error)
}

type gatewayRefreshTokenServiceImpl struct {
	cache                 cassandracache.Cacher
	jwtProvider           jwt.JWTProvider
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase
}

func NewGatewayRefreshTokenService(
	cach cassandracache.Cacher,
	jwtp jwt.JWTProvider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
) GatewayRefreshTokenService {
	return &gatewayRefreshTokenServiceImpl{cach, jwtp, uc1}
}

type GatewayRefreshTokenRequestIDO struct {
	Value string `json:"value"`
}

// GatewayRefreshTokenResponseIDO struct used to represent the system's response when the `login` POST request was a success.
type GatewayRefreshTokenResponseIDO struct {
	Email                  string    `json:"username"`
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
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
	atExpiry := 5 * time.Minute     // 5 minutes
	rtExpiry := 14 * 24 * time.Hour // 1 week

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

	ido := &GatewayRefreshTokenResponseIDO{
		Email:                  u.Email,
		AccessToken:            accessToken,
		AccessTokenExpiryDate:  accessTokenExpiry,
		RefreshToken:           refreshToken,
		RefreshTokenExpiryDate: refreshTokenExpiry,
	}

	// Return our auth keys.
	return ido, nil
}
