// monorepo/native/desktop/maplefile-cli/internal/service/authdto/completelogin.go
package authdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	uc_medto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/medto"
)

// CompleteLoginService provides high-level functionality for login completion
type CompleteLoginService interface {
	CompleteLogin(ctx context.Context, email, password string) (*dom_authdto.TokenResponseDTO, error)
}

// completeLoginService implements the CompleteLoginService interface
type completeLoginService struct {
	logger                 *zap.Logger
	useCase                uc_authdto.CompleteLoginUseCase
	userRepo               user.Repository
	configService          config.ConfigService
	getMeFromCloudUseCase  uc_medto.GetMeFromCloudUseCase
	tokenDecryptionService TokenDecryptionService
}

// NewCompleteLoginService creates a new login completion service
func NewCompleteLoginService(
	logger *zap.Logger,
	useCase uc_authdto.CompleteLoginUseCase,
	userRepo user.Repository,
	configService config.ConfigService,
	getMeFromCloudUseCase uc_medto.GetMeFromCloudUseCase,
	tokenDecryptionService TokenDecryptionService,
) CompleteLoginService {
	logger = logger.Named("CompleteLoginService")
	return &completeLoginService{
		logger:                 logger,
		useCase:                useCase,
		userRepo:               userRepo,
		configService:          configService,
		getMeFromCloudUseCase:  getMeFromCloudUseCase,
		tokenDecryptionService: tokenDecryptionService,
	}
}

// CompleteLogin handles the entire flow of login completion
func (s *completeLoginService) CompleteLogin(ctx context.Context, email, password string) (*dom_authdto.TokenResponseDTO, error) {
	// Call the use case to complete login and get token and updated user
	tokenResp, updatedUser, err := s.useCase.CompleteLogin(ctx, email, password)
	if err != nil {
		return nil, errors.NewAppError("failed to complete login", err)
	}

	// All tokens should now be encrypted - verify we received encrypted tokens
	if tokenResp.EncryptedTokens == "" {
		return nil, errors.NewAppError("server did not return encrypted tokens - this should not happen in encrypted-only mode", nil)
	}

	s.logger.Info("Received encrypted tokens, decrypting...",
		zap.String("email", email))

	// Decrypt the tokens using the user's private key
	accessToken, refreshToken, err := s.tokenDecryptionService.DecryptTokens(
		tokenResp.EncryptedTokens,
		updatedUser,
		password,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt authentication tokens", err)
	}

	// Update the response with decrypted tokens
	tokenResp.AccessToken = accessToken
	tokenResp.RefreshToken = refreshToken

	s.logger.Info("Successfully decrypted authentication tokens",
		zap.String("email", email))

	// Start a transaction to update the user
	if err := s.userRepo.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}

	// Save the updated user
	if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to update user data", err)
	}

	// Save our authenticated credentials to configuration
	s.configService.SetLoggedInUserCredentials(
		ctx,
		email,
		tokenResp.AccessToken,
		&tokenResp.AccessTokenExpiryTime,
		tokenResp.RefreshToken,
		&tokenResp.RefreshTokenExpiryTime,
	)

	// Log success
	s.logger.Info("✅ Login completed successfully with encrypted tokens",
		zap.String("email", email),
		zap.Time("accessTokenExpiry", tokenResp.AccessTokenExpiryTime),
		zap.Time("refreshTokenExpiry", tokenResp.RefreshTokenExpiryTime))

	// Fetch the profile
	meDTO, err := s.getMeFromCloudUseCase.Execute(ctx)
	if err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to get user profile post successful complete login", err)
	}
	if meDTO == nil {
		s.logger.Error("❌ Failed to get user profile from cloud because backend returned nil")
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to get user profile from cloud because backend returned nil", nil)
	}

	// Update the local user profile from the cloud
	updatedUser.ID = meDTO.ID
	updatedUser.FirstName = meDTO.FirstName
	updatedUser.LastName = meDTO.LastName
	updatedUser.Name = meDTO.Name
	updatedUser.LexicalName = meDTO.LexicalName
	updatedUser.Role = meDTO.Role
	updatedUser.WasEmailVerified = meDTO.WasEmailVerified
	updatedUser.Phone = meDTO.Phone
	updatedUser.Country = meDTO.Country
	updatedUser.Timezone = meDTO.Timezone
	updatedUser.Region = meDTO.Region
	updatedUser.City = meDTO.City
	updatedUser.PostalCode = meDTO.PostalCode
	updatedUser.AddressLine1 = meDTO.AddressLine1
	updatedUser.AddressLine2 = meDTO.AddressLine2
	updatedUser.AgreePromotions = meDTO.AgreePromotions
	updatedUser.AgreeToTrackingAcrossThirdPartyAppsAndServices = meDTO.AgreeToTrackingAcrossThirdPartyAppsAndServices
	updatedUser.CreatedAt = meDTO.CreatedAt
	updatedUser.Status = meDTO.Status

	if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to update user data", err)
	}

	// Commit the transaction
	if err := s.userRepo.CommitTransaction(); err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	// Log success
	s.logger.Info("✅ Successfully received user profile and saved locally with encrypted tokens",
		zap.String("email", email),
		zap.Time("accessTokenExpiry", tokenResp.AccessTokenExpiryTime),
		zap.Time("refreshTokenExpiry", tokenResp.RefreshTokenExpiryTime))

	return tokenResp, nil
}
