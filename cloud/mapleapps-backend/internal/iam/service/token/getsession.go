package token

import (
	"context"
	"time"

	"go.uber.org/zap"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/database/mongodbcache"
)

type TokenGetSessionService interface {
	Execute(sessCtx context.Context, sessionID string) (*dom_user.FederatedUser, error)
}

type tokenGetSessionServiceImpl struct {
	logger                    *zap.Logger
	passwordProvider          password.Provider
	cache                     mongodbcache.Cacher
	jwtProvider               jwt.Provider
	userGetBySessionIDUseCase uc_user.FederatedUserGetBySessionIDUseCase
}

func NewTokenGetSessionService(
	logger *zap.Logger,
	pp password.Provider,
	cach mongodbcache.Cacher,
	jwtp jwt.Provider,
	userGetBySessionIDUseCase uc_user.FederatedUserGetBySessionIDUseCase,
) TokenGetSessionService {
	return &tokenGetSessionServiceImpl{logger, pp, cach, jwtp, userGetBySessionIDUseCase}
}

type TokenGetSessionRequestIDO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenGetSessionResponseIDO struct {
	FederatedUser          *domain.FederatedUser `json:"federateduser"`
	AccessToken            string                `json:"access_token"`
	AccessTokenExpiryTime  time.Time             `json:"access_token_expiry_time"`
	RefreshToken           string                `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time             `json:"refresh_token_expiry_time"`
}

func (impl *tokenGetSessionServiceImpl) Execute(ctx context.Context, sessionID string) (*dom_user.FederatedUser, error) {
	// Lookup our user profile in the session or return 500 error.
	user, err := impl.userGetBySessionIDUseCase.Execute(ctx, sessionID)
	if err != nil {
		impl.logger.Warn("GetFederatedUserBySessionID error", zap.Any("err", err), zap.Any("middleware", "PostJWTProcessorMiddleware"))
		return nil, httperror.NewForUnauthorizedWithSingleField("message", "error looking up session id")
	}

	// If no user was found then that means our session expired and the
	// user needs to login or use the refresh token.
	if user == nil {
		impl.logger.Warn("Session expired - please log in again", zap.Any("middleware", "PostJWTProcessorMiddleware"))
		return nil, httperror.NewForUnauthorizedWithSingleField("message", "attempting to access a protected endpoin")
	}

	// // If system administrator disabled the user account then we need
	// // to generate a 403 error letting the user know their account has
	// // been disabled and you cannot access the protected API endpoint.
	// if user.State == 0 {
	// 	http.Error(w, "Account disabled - please contact admin", http.StatusForbidden)
	// 	return
	// }

	// For debugging purposes only.
	impl.logger.Debug("Fetched session record",
		zap.Any("ID", user.ID),
		zap.String("SessionID", sessionID),
		zap.String("Name", user.Name),
		zap.String("FirstName", user.FirstName),
		zap.String("Email", user.Email))

	return user, nil
}
