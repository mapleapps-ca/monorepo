// monorepo/native/desktop/maplefile-cli/internal/usecase/auth/verifyloginott.go
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// LoginOTTVerificationUseCase defines the interface for login OTT verification use cases
type LoginOTTVerificationUseCase interface {
	VerifyLoginOTT(ctx context.Context, email, ott string) (*dom_authdto.VerifyLoginOTTResponse, *user.User, error)
}

// loginOTTVerificationUseCase implements the LoginOTTVerificationUseCase interface
type loginOTTVerificationUseCase struct {
	logger          *zap.Logger
	repository      dom_authdto.LoginOTTVerificationRepository
	userRepo        user.Repository
	dataTransformer dom_authdto.UserVerificationDataTransformer
}

// NewLoginOTTVerificationUseCase creates a new login OTT verification use case
func NewLoginOTTVerificationUseCase(
	logger *zap.Logger,
	repository dom_authdto.LoginOTTVerificationRepository,
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
func (uc *loginOTTVerificationUseCase) VerifyLoginOTT(ctx context.Context, email, ott string) (*dom_authdto.VerifyLoginOTTResponse, *user.User, error) {
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
		//TODO: REMOVE THIS
		return nil, nil, errors.NewAppError(
			fmt.Sprintf("user with email %s not found; please register first", email),
			nil,
		)
	}

	// Create request and verify the OTT
	request := &dom_authdto.VerifyLoginOTTRequest{
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

	return response, existingUser, nil
}
