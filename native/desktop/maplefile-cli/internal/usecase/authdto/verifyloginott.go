// monorepo/native/desktop/maplefile-cli/internal/usecase/authdto/verifyloginott.go
package authdto

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// LoginOTTVerificationUseCase defines the interface for login OTT verification use cases
type LoginOTTVerificationUseCase interface {
	VerifyLoginOTT(ctx context.Context, email, ott string) (*dom_authdto.VerifyLoginOTTResponseDTO, *user.User, error)
}

// loginOTTVerificationUseCase implements the LoginOTTVerificationUseCase interface
type loginOTTVerificationUseCase struct {
	logger          *zap.Logger
	repository      dom_authdto.LoginOTTVerificationDTORepository
	userRepo        user.Repository
	dataTransformer dom_authdto.UserVerificationDataTransformer
}

// NewLoginOTTVerificationUseCase creates a new login OTT verification use case
func NewLoginOTTVerificationUseCase(
	logger *zap.Logger,
	repository dom_authdto.LoginOTTVerificationDTORepository,
	userRepo user.Repository,
	dataTransformer dom_authdto.UserVerificationDataTransformer,
) LoginOTTVerificationUseCase {
	logger = logger.Named("LoginOTTVerificationUseCase")
	return &loginOTTVerificationUseCase{
		logger:          logger,
		repository:      repository,
		userRepo:        userRepo,
		dataTransformer: dataTransformer,
	}
}

// VerifyLoginOTT verifies a login OTT and updates the user with verification data
func (uc *loginOTTVerificationUseCase) VerifyLoginOTT(ctx context.Context, email, ott string) (*dom_authdto.VerifyLoginOTTResponseDTO, *user.User, error) {
	// Validate inputs
	if email == "" {
		return nil, nil, errors.NewAppError("email is required", nil)
	}
	if ott == "" {
		return nil, nil, errors.NewAppError("one-time token is required", nil)
	}

	// Sanitize inputs
	email = strings.ToLower(strings.TrimSpace(email))
	ott = strings.TrimSpace(ott)

	// Check if user exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to retrieve user", err)
	}

	if existingUser == nil {
		uc.logger.Debug("no existing user exists, create lite user now",
			zap.String("email", email),
		)
		newUser := &user.User{
			Email: email,
		}

		if err := uc.userRepo.UpsertByEmail(ctx, newUser); err != nil {
			return nil, nil, errors.NewAppError("failed to save new user", err)
		}

		// Now our new user exists in our database so return it now.
		existingUser = newUser

		uc.logger.Debug("new user exists on local database",
			zap.String("email", email),
		)
	}

	// Create request and verify the OTT
	request := &dom_authdto.VerifyLoginOTTRequestDTO{
		Email: email,
		OTT:   ott,
	}

	response, err := uc.repository.VerifyLoginOTT(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	// Update user with verification data
	if err := uc.dataTransformer.UpdateUserWithVerificationData(existingUser, response); err != nil {
		return nil, nil, errors.NewAppError("failed to update user with verification data", err)
	}

	// Update modified timestamp
	existingUser.ModifiedAt = time.Now()

	if err := uc.userRepo.UpsertByEmail(ctx, existingUser); err != nil {
		return nil, nil, errors.NewAppError("failed to save existing user", err)
	}

	uc.logger.Debug("existing user updated with local keys",
		zap.String("email", email),
	)

	return response, existingUser, nil
}
