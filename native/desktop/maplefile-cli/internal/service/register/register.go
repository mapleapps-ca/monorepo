// internal/service/register/register.go
package register

import (
	"context"
	"fmt"
	"strings"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	registerUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/register"
	userUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// RegisterUserInput contains the inputs required for user registration
type RegisterUserInput struct {
	Email           string
	Password        string
	FirstName       string
	LastName        string
	Timezone        string
	Country         string
	Phone           string
	BetaAccessCode  string
	AgreeTerms      bool
	AgreePromotions bool
	AgreeTracking   bool
	Module          int
	SkipRemoteReg   bool
}

// RegisterUserOutput contains the output of the registration process
type RegisterUserOutput struct {
	User                 *user.User
	RecoveryKey          []byte
	RegistrationComplete bool
	ServerResponse       string
}

// RegisterService defines the interface for registration service
type RegisterService interface {
	RegisterUser(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error)
}

// registerService implements the RegisterService interface
type registerService struct {
	txManager                       transaction.Manager
	generateCredentialsUseCase      registerUseCase.GenerateCredentialsUseCase
	createLocalUserUseCase          registerUseCase.CreateLocalUserUseCase
	upsertUserByEmailUseCase        userUseCase.UpsertByEmailUseCase
	sendRegistrationToServerUseCase registerUseCase.SendRegistrationToServerUseCase
}

// NewRegisterService creates a new instance of RegisterService
func NewRegisterService(
	txManager transaction.Manager,
	generateCredentialsUseCase registerUseCase.GenerateCredentialsUseCase,
	createLocalUserUseCase registerUseCase.CreateLocalUserUseCase,
	upsertUserByEmailUseCase userUseCase.UpsertByEmailUseCase,
	sendRegistrationToServerUseCase registerUseCase.SendRegistrationToServerUseCase,
) RegisterService {
	return &registerService{
		txManager:                       txManager,
		generateCredentialsUseCase:      generateCredentialsUseCase,
		createLocalUserUseCase:          createLocalUserUseCase,
		upsertUserByEmailUseCase:        upsertUserByEmailUseCase,
		sendRegistrationToServerUseCase: sendRegistrationToServerUseCase,
	}
}

// RegisterUser handles the complete registration process
func (s *registerService) RegisterUser(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	// Generate E2EE credentials
	credentials, err := s.generateCredentialsUseCase.Execute(ctx, input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Create local user
	userInput := registerUseCase.CreateLocalUserInput{
		Email:           strings.ToLower(input.Email),
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Timezone:        input.Timezone,
		Country:         input.Country,
		Phone:           input.Phone,
		AgreeTerms:      input.AgreeTerms,
		AgreePromotions: input.AgreePromotions,
		AgreeTracking:   input.AgreeTracking,
		Credentials:     credentials,
	}

	newUser, err := s.createLocalUserUseCase.Execute(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Start a transaction to store the user
	if err := s.txManager.Begin(); err != nil {
		return nil, fmt.Errorf("error opening transaction: %w", err)
	}

	// Store the user
	if err := s.upsertUserByEmailUseCase.Execute(ctx, newUser); err != nil {
		s.txManager.Rollback()
		return nil, fmt.Errorf("error saving user data: %w", err)
	}

	output := &RegisterUserOutput{
		User:                 newUser,
		RecoveryKey:          credentials.RecoveryKey,
		RegistrationComplete: true,
	}

	// Optionally send registration to server
	if !input.SkipRemoteReg {
		serverInput := registerUseCase.SendRegistrationToServerInput{
			User:           newUser,
			BetaAccessCode: input.BetaAccessCode,
			Module:         input.Module,
		}

		response, err := s.sendRegistrationToServerUseCase.Execute(ctx, serverInput)
		if err != nil {
			s.txManager.Rollback()
			return nil, fmt.Errorf("failed to send registration to server: %w", err)
		}
		output.ServerResponse = response
	}

	// Commit the transaction
	if err := s.txManager.Commit(); err != nil {
		s.txManager.Rollback()
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return output, nil
}
