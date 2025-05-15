// internal/usecase/register/generate_credentials.go
package register

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/nacl/box"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
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

	// Generate key pair
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}
	publicKey := pubKey[:]
	privateKey := privKey[:]

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

	// Create verification ID from public key
	verificationID := base64.URLEncoding.EncodeToString(publicKey)[:12]

	return &Credentials{
		Salt:             salt,
		MasterKey:        masterKey,
		KeyEncryptionKey: keyEncryptionKey,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		RecoveryKey:      recoveryKey,
		EncryptedMasterKey: keys.EncryptedMasterKey{
			Ciphertext: encryptedMasterKey.Ciphertext,
			Nonce:      encryptedMasterKey.Nonce,
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
