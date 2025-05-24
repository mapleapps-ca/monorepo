// native/desktop/maplefile-cli/internal/usecase/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
)

// RefreshTokenUseCase defines the interface for refreshing tokens
type RefreshTokenUseCase interface {
	Execute(ctx context.Context) error
}

// refreshTokenUseCaseImpl implements the RefreshTokenUseCase interface
type refreshTokenUseCaseImpl struct {
	logger          *zap.Logger
	tokenRepository auth.TokenRepository
}

// NewRefreshTokenUseCase creates a new instance of RefreshTokenUseCase
func NewRefreshTokenUseCase(
	logger *zap.Logger,
	tokenRepository auth.TokenRepository,
) RefreshTokenUseCase {
	return &refreshTokenUseCaseImpl{
		logger:          logger,
		tokenRepository: tokenRepository,
	}
}

// Execute performs the token refresh operation
func (uc *refreshTokenUseCaseImpl) Execute(ctx context.Context) error {
	_, err := uc.tokenRepository.GetAccessTokenAfterForcedRefresh(ctx)
	if err != nil {
		return fmt.Errorf("error getting authenticated user access token after forced refresh: %w", err)
	}
	return nil
}
