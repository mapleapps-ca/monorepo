// internal/usecase/register/generate_credentials.go
package register

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// Credentials contains the generated cryptographic credentials
type Credentials struct {
	Salt                              []byte
	MasterKey                         []byte
	KeyEncryptionKey                  []byte
	PublicKey                         []byte
	PrivateKey                        []byte
	RecoveryKey                       []byte
	EncryptedMasterKey                keys.EncryptedMasterKey
	EncryptedPrivateKey               keys.EncryptedPrivateKey
	EncryptedRecoveryKey              keys.EncryptedRecoveryKey
	MasterKeyEncryptedWithRecoveryKey keys.MasterKeyEncryptedWithRecoveryKey
	VerificationID                    string
}

// GenerateCredentialsUseCase defines the interface for generating credentials
type GenerateCredentialsUseCase interface {
	Execute(ctx context.Context, password string) (*Credentials, error)
}

type generateCredentialsUseCase struct{}

// NewGenerateCredentialsUseCase creates a new GenerateCredentialsUseCase
func NewGenerateCredentialsUseCase() GenerateCredentialsUseCase {
	return &generateCredentialsUseCase{}
}

// Execute generates credentials for a user
func (uc *generateCredentialsUseCase) Execute(ctx context.Context, password string) (*Credentials, error) {
	// Generate salt for key derivation
	salt, err := crypto.GenerateRandomBytes(crypto.Argon2SaltSize)
	if err != nil {
		return nil, fmt.Errorf("error generating salt: %w", err)
	}

	// Derive key from password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, salt)
	if err != nil {
		return nil, fmt.Errorf("error deriving key from password: %w", err)
	}

	// Generate master key
	masterKey, err := crypto.GenerateRandomBytes(crypto.MasterKeySize)
	if err != nil {
		return nil, fmt.Errorf("error generating master key: %w", err)
	}

	// Generate key pair (and create verification ID from public key too!).
	publicKey, privateKey, verificationID, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	log.Printf("usecase.register.GenerateCredentialsUseCase - publicKey: %v\n", publicKey)

	// Defensive Coding: Verify the `verificationID` to make sure we are enforcing the fact that `verificationID` are derived from public keys.
	if !crypto.VerifyVerificationID(publicKey, verificationID) {
		return nil, fmt.Errorf("failed to verify the user verification id, was it generated from the public key?")
	}

	// Generate recovery key
	recoveryKey, err := crypto.GenerateRandomBytes(crypto.RecoveryKeySize)
	if err != nil {
		return nil, fmt.Errorf("error generating recovery key: %w", err)
	}

	// Encrypt master key with key encryption key
	encryptedMasterKey, err := crypto.EncryptWithSecretBox(masterKey, keyEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting master key: %w", err)
	}

	// Encrypt private key with master key
	encryptedPrivateKey, err := crypto.EncryptWithSecretBox(privateKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting private key: %w", err)
	}

	// Encrypt recovery key with master key
	encryptedRecoveryKey, err := crypto.EncryptWithSecretBox(recoveryKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting recovery key: %w", err)
	}

	// Encrypt master key with recovery key
	masterKeyEncryptedWithRecoveryKey, err := crypto.EncryptWithSecretBox(masterKey, recoveryKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting master key with recovery key: %w", err)
	}

	// Store current key in history
	currentTime := time.Now()
	historicalKey := keys.EncryptedHistoricalKey{
		KeyVersion:    1,
		Ciphertext:    encryptedMasterKey.Ciphertext,
		Nonce:         encryptedMasterKey.Nonce,
		RotatedAt:     currentTime,
		RotatedReason: "Initial user registration",
		Algorithm:     "chacha20poly1305",
	}

	return &Credentials{
		Salt:             salt,
		MasterKey:        masterKey,
		KeyEncryptionKey: keyEncryptionKey,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		RecoveryKey:      recoveryKey,
		EncryptedMasterKey: keys.EncryptedMasterKey{
			Ciphertext:   encryptedMasterKey.Ciphertext,
			Nonce:        encryptedMasterKey.Nonce,
			KeyVersion:   1,
			RotatedAt:    &currentTime,
			PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
		},
		EncryptedPrivateKey: keys.EncryptedPrivateKey{
			Ciphertext: encryptedPrivateKey.Ciphertext,
			Nonce:      encryptedPrivateKey.Nonce,
		},
		EncryptedRecoveryKey: keys.EncryptedRecoveryKey{
			Ciphertext: encryptedRecoveryKey.Ciphertext,
			Nonce:      encryptedRecoveryKey.Nonce,
		},
		MasterKeyEncryptedWithRecoveryKey: keys.MasterKeyEncryptedWithRecoveryKey{
			Ciphertext: masterKeyEncryptedWithRecoveryKey.Ciphertext,
			Nonce:      masterKeyEncryptedWithRecoveryKey.Nonce,
		},
		VerificationID: verificationID,
	}, nil
}
