// internal/usecase/register/create_local_user.go
package register

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
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
	// Create a new user entity
	newUser := &user.User{
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

		// E2EE related fields
		PasswordSalt:       input.Credentials.Salt,
		EncryptedMasterKey: input.Credentials.EncryptedMasterKey,
		PublicKey: keys.PublicKey{
			Key:            input.Credentials.PublicKey,
			VerificationID: input.Credentials.VerificationID,
		},
		EncryptedPrivateKey:               input.Credentials.EncryptedPrivateKey,
		EncryptedRecoveryKey:              input.Credentials.EncryptedRecoveryKey,
		MasterKeyEncryptedWithRecoveryKey: input.Credentials.MasterKeyEncryptedWithRecoveryKey,
		VerificationID:                    input.Credentials.VerificationID,

		// KDFParams related
		KDFParams: dom_user.KDFParams{
			Algorithm:   crypto.Argon2IDAlgorithm,
			Iterations:  crypto.Argon2OpsLimit,
			Memory:      crypto.Argon2MemLimit,
			Parallelism: crypto.Argon2Parallelism,
		},

		// Terms agreements
		AgreeTermsOfService: input.AgreeTerms,
		AgreePromotions:     input.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: input.AgreeTracking,

		// Metadata
		CreatedAt: time.Now(),
	}

	return newUser, nil
}
