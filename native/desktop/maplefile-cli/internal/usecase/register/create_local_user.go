// internal/usecase/register/create_local_user.go
package register

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// CreateLocalUserInput contains the inputs for creating a user
type CreateLocalUserInput struct {
	Email           string
	FirstName       string
	LastName        string
	Timezone        string
	Country         string
	Phone           string
	AgreeTerms      bool
	AgreePromotions bool
	AgreeTracking   bool
	Credentials     *Credentials
}

// CreateLocalUserUseCase defines the interface for creating a local user
type CreateLocalUserUseCase interface {
	Execute(ctx context.Context, input CreateLocalUserInput) (*user.User, error)
}

type createLocalUserUseCase struct{}

// NewCreateLocalUserUseCase creates a new CreateLocalUserUseCase
func NewCreateLocalUserUseCase() CreateLocalUserUseCase {
	return &createLocalUserUseCase{}
}

// Execute creates a new local user
func (uc *createLocalUserUseCase) Execute(ctx context.Context, input CreateLocalUserInput) (*user.User, error) {
	currentTime := time.Now() // Capture the current time once

	// Create a new user entity
	newUser := &user.User{
		// E2EE related fields
		PasswordSalt:       input.Credentials.Salt,
		KDFParams:          keys.DefaultKDFParams(),
		EncryptedMasterKey: input.Credentials.EncryptedMasterKey,
		PublicKey: keys.PublicKey{
			Key:            input.Credentials.PublicKey,
			VerificationID: input.Credentials.VerificationID,
		},
		EncryptedPrivateKey:               input.Credentials.EncryptedPrivateKey,
		EncryptedRecoveryKey:              input.Credentials.EncryptedRecoveryKey,
		MasterKeyEncryptedWithRecoveryKey: input.Credentials.MasterKeyEncryptedWithRecoveryKey,
		VerificationID:                    input.Credentials.VerificationID,
		LastPasswordChange:                time.Now(),
		KDFParamsNeedUpgrade:              false,
		CurrentKeyVersion:                 1,
		LastKeyRotation:                   &currentTime,
		KeyRotationPolicy:                 nil,

		// --- The rest of the stuff... ---
		ID:               primitive.NewObjectID(),
		Email:            input.Email,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Name:             input.FirstName + " " + input.LastName,
		LexicalName:      input.LastName + ", " + input.FirstName,
		Role:             user.UserRoleIndividual,
		Status:           user.UserStatusActive,
		WasEmailVerified: false,
		Phone:            input.Phone,
		Country:          input.Country,
		Timezone:         input.Timezone,

		// Terms agreements
		AgreeTermsOfService: input.AgreeTerms,
		AgreePromotions:     input.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: input.AgreeTracking,

		// Metadata
		CreatedAt: time.Now(),
	}

	return newUser, nil
}
