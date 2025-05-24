// monorepo/native/desktop/maplefile-cli/internal/usecase/auth/completelogin.go
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// CompleteLoginUseCase defines the interface for login completion use cases
type CompleteLoginUseCase interface {
	CompleteLogin(ctx context.Context, email, password string) (*auth.TokenResponse, *user.User, error)
}

// completeLoginUseCase implements the CompleteLoginUseCase interface
type completeLoginUseCase struct {
	logger          *zap.Logger
	tokenRepository auth.TokenRepository
	repository      auth.CompleteLoginRepository
	userRepo        user.Repository
	cryptoService   auth.CryptographyOperations
}

// NewCompleteLoginUseCase creates a new login completion use case
func NewCompleteLoginUseCase(
	logger *zap.Logger,
	tokenRepository auth.TokenRepository,
	repository auth.CompleteLoginRepository,
	userRepo user.Repository,
	cryptoService auth.CryptographyOperations,
) CompleteLoginUseCase {
	return &completeLoginUseCase{
		logger:          logger,
		tokenRepository: tokenRepository,
		repository:      repository,
		userRepo:        userRepo,
		cryptoService:   cryptoService,
	}
}

// CompleteLogin handles the business logic for login completion
func (uc *completeLoginUseCase) CompleteLogin(ctx context.Context, email, password string) (*auth.TokenResponse, *user.User, error) {
	// Validate inputs
	if email == "" {
		return nil, nil, errors.NewAppError("email is required", nil)
	}
	if password == "" {
		return nil, nil, errors.NewAppError("password is required", nil)
	}

	// Sanitize inputs
	email = strings.ToLower(strings.TrimSpace(email))

	// Get user from repository
	userData, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to retrieve user data", err)
	}

	if userData == nil {
		return nil, nil, errors.NewAppError(fmt.Sprintf("user with email %s not found", email), nil)
	}

	// Get challenge ID
	challengeID := userData.VerificationID
	if challengeID == "" {
		return nil, nil, errors.NewAppError("no challenge ID found; please run verifyloginott first", nil)
	}

	uc.logger.Debug("Processing login completion",
		zap.String("email", email),
		zap.String("challengeID", challengeID),
		zap.Int("salt length", len(userData.PasswordSalt)),
		zap.Int("public key length", len(userData.PublicKey.Key)),
		zap.Int("encrypted challenge length", len(userData.EncryptedChallenge)))

	// Derive key from password and salt
	keyEncryptionKey, err := uc.cryptoService.DeriveKeyFromPassword(password, userData.PasswordSalt)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to derive key from password", err)
	}

	// Decrypt Master Key using Key Encryption Key
	masterKey, err := uc.cryptoService.DecryptWithSecretBox(
		userData.EncryptedMasterKey.Ciphertext,
		userData.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to decrypt master key", err)
	}

	// Decrypt Private Key using Master Key
	privateKey, err := uc.cryptoService.DecryptWithSecretBox(
		userData.EncryptedPrivateKey.Ciphertext,
		userData.EncryptedPrivateKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to decrypt private key", err)
	}

	// Decrypt Challenge using Public and Private Keys
	decryptedChallenge, err := uc.cryptoService.DecryptWithBox(
		userData.EncryptedChallenge,
		userData.PublicKey.Key,
		privateKey,
	)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to decrypt challenge", err)
	}

	// Encode decrypted challenge to base64
	decryptedChallengeBase64 := uc.cryptoService.EncodeToBase64(decryptedChallenge)

	// Send decrypted challenge to server to complete login
	completeLoginReq := &auth.CompleteLoginRequest{
		Email:         email,
		ChallengeID:   challengeID,
		DecryptedData: decryptedChallengeBase64,
	}

	tokenResp, err := uc.repository.CompleteLogin(ctx, completeLoginReq)
	if err != nil {
		return nil, nil, err
	}

	// Update user metadata.
	userData.LastLoginAt = time.Now()
	userData.ModifiedAt = time.Now()
	if err := uc.userRepo.UpsertByEmail(ctx, userData); err != nil {
		return nil, nil, err
	}

	// IMPORTANT: Save the authenticated user credentials to our token repository.
	// DEVELOPER NOTE: The token repository is essentially utilizing the `config` package to store the credentials to the applications data folder.
	uc.tokenRepository.Save(
		ctx,
		email,
		tokenResp.AccessToken,
		&tokenResp.AccessTokenExpiryTime,
		tokenResp.RefreshToken,
		&tokenResp.RefreshTokenExpiryTime)

	return tokenResp, userData, nil
}
