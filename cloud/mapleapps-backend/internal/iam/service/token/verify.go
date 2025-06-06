package token

import (
	"context"
	"strings"

	"go.uber.org/zap"

	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/database/mongodbcache"
)

type TokenVerifyService interface {
	Execute(sessCtx context.Context, token string) (string, error)
}

type tokenVerifyServiceImpl struct {
	logger           *zap.Logger
	passwordProvider password.Provider
	cache            mongodbcache.Cacher
	jwtProvider      jwt.Provider
}

func NewTokenVerifyService(
	logger *zap.Logger,
	pp password.Provider,
	cach mongodbcache.Cacher,
	jwtp jwt.Provider,
	userGetBySessionIDUseCase uc_user.FederatedUserGetBySessionIDUseCase) TokenVerifyService {
	return &tokenVerifyServiceImpl{logger, pp, cach, jwtp}
}

func (impl *tokenVerifyServiceImpl) Execute(ctx context.Context, rawToken string) (string, error) {
	// For debugging purposes, using the logger from mid.
	impl.logger.Debug("Authorization header received",
		zap.String("Authorization", rawToken))

	// Check if the Authorization header is present and valid.
	if rawToken == "" || strings.Contains(strings.ToLower(rawToken), "undefined") {
		impl.logger.Warn("authorization header missing, empty, or contains 'undefined'",
			zap.String("header_value", rawToken))
		// http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return "", httperror.NewForUnauthorizedWithSingleField("message", "authorization header missing, empty, or contains 'undefined'")
	}

	// Expecting "Bearer <token>" or "JWT <token>" format.
	// Standard is typically "Bearer", but the original code used "JWT ".
	// Let's stick to "JWT " as per the original logic for now.
	const prefix = "JWT "
	if !strings.HasPrefix(rawToken, prefix) {
		impl.logger.Warn("authorization header not properly formatted",
			zap.String("header_value", rawToken))
		return "", httperror.NewForUnauthorizedWithSingleField("message", "authorization header format must be 'JWT token'")
	}

	// Extract the actual token string.
	actualToken := strings.TrimPrefix(rawToken, prefix)
	if actualToken == "" {
		impl.logger.Warn("authorization token is empty after stripping prefix",
			zap.String("header_value", rawToken))
		return "", httperror.NewForUnauthorizedWithSingleField("message", "authorization token cannot be empty")
	}

	// Accessing the JWT service from the mid struct to process the token.
	sessionID, err := impl.jwtProvider.ProcessJWTToken(actualToken)
	if err != nil {
		// Log the specific error from JWT processing for better debugging.
		impl.logger.Error("failed to process JWT token",
			zap.Error(err))
		// Return a generic error message to the client.
		return "", httperror.NewForUnauthorizedWithSingleField("message", "Invalid or expired token")
	}

	return sessionID, nil
}
