// cloud/backend/internal/iam/service/gateway/requestott.go
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	uc_emailer "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/emailer"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/random"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodbcache"
)

// Data structures for OTT request
type GatewayRequestLoginOTTRequestIDO struct {
	Email string `json:"email"`
}

type GatewayRequestLoginOTTResponseIDO struct {
	Message string `json:"message"`
}

// Service interface for OTT request
type GatewayRequestLoginOTTService interface {
	Execute(sessCtx context.Context, req *GatewayRequestLoginOTTRequestIDO) (*GatewayRequestLoginOTTResponseIDO, error)
}

// Login OTT data structure (to be stored in cache)
type LoginOTTData struct {
	Email       string    `json:"email"`
	OTT         string    `json:"ott"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ClientIP    string    `json:"client_ip"`
	IsVerified  bool      `json:"is_verified"`
	ChallengeID string    `json:"challenge_id,omitempty"`
}

// Implementation of OTT request service
type gatewayRequestLoginOTTServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	cache                 mongodbcache.Cacher
	jwtProvider           jwt.Provider
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase
	sendOTTEmailUseCase   uc_emailer.SendLoginOTTEmailUseCase
}

func NewGatewayRequestLoginOTTService(
	config *config.Configuration,
	logger *zap.Logger,
	cache mongodbcache.Cacher,
	jwtProvider jwt.Provider,
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase,
	sendOTTEmailUseCase uc_emailer.SendLoginOTTEmailUseCase,
) GatewayRequestLoginOTTService {
	return &gatewayRequestLoginOTTServiceImpl{
		config:                config,
		logger:                logger,
		cache:                 cache,
		jwtProvider:           jwtProvider,
		userGetByEmailUseCase: userGetByEmailUseCase,
		sendOTTEmailUseCase:   sendOTTEmailUseCase,
	}
}

func (s *gatewayRequestLoginOTTServiceImpl) Execute(sessCtx context.Context, req *GatewayRequestLoginOTTRequestIDO) (*GatewayRequestLoginOTTResponseIDO, error) {
	// Validate input
	e := make(map[string]string)
	if req.Email == "" {
		e["email"] = "Email address is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Sanitize input
	req.Email = strings.ToLower(req.Email)
	req.Email = strings.ReplaceAll(req.Email, " ", "")

	// Check if user exists
	user, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exist")
	}

	// Generate OTT
	ott, err := random.GenerateSixDigitCode()
	if err != nil {
		return nil, err
	}

	// Store OTT in cache
	ottData := LoginOTTData{
		Email:      req.Email,
		OTT:        ott,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute), // OTT valid for 10 minutes
		ClientIP:   "",                               // IP could be stored here if needed
		IsVerified: false,
	}

	// Generate a unique cache key for this OTT
	cacheKey := fmt.Sprintf("login_ott:%s", req.Email)

	// Marshal the data to JSON
	ottDataJSON, err := json.Marshal(ottData)
	if err != nil {
		s.logger.Error("Failed to marshal OTT data", zap.Error(err))
		return nil, fmt.Errorf("failed to process login request: %w", err)
	}

	// Store in cache with expiry
	if err := s.cache.SetWithExpiry(sessCtx, cacheKey, ottDataJSON, 10*time.Minute); err != nil {
		s.logger.Error("Failed to store OTT in cache", zap.Error(err))
		return nil, fmt.Errorf("failed to process login request: %w", err)
	}

	// Send OTT via email
	// TODO: Replace `int(constants.MonolithModuleMapleFile)` with int value from request payload.
	if err := s.sendOTTEmailUseCase.Execute(sessCtx, int(constants.MonolithModuleMapleFile), user.Email, ott, user.FirstName); err != nil {
		s.logger.Error("Failed to send OTT email", zap.Error(err))
		return nil, fmt.Errorf("failed to send login code: %w", err)
	}

	return &GatewayRequestLoginOTTResponseIDO{
		Message: "A verification code has been sent to your email",
	}, nil
}
