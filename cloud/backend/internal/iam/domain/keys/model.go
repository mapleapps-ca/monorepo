package keys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// MasterKey represents the root encryption key for a user
type MasterKey struct {
	Key []byte `json:"key" bson:"key"`
}

// EncryptedMasterKey is the master key encrypted with the key encryption key
type EncryptedMasterKey struct {
	Ciphertext   []byte                   `json:"ciphertext" bson:"ciphertext"`
	Nonce        []byte                   `json:"nonce" bson:"nonce"`
	KeyVersion   int                      `json:"key_version" bson:"key_version"`
	RotatedAt    *time.Time               `json:"rotated_at,omitempty" bson:"rotated_at,omitempty"`
	PreviousKeys []EncryptedHistoricalKey `json:"previous_keys,omitempty" bson:"previous_keys,omitempty"`
}

func (emk *EncryptedMasterKey) GetCurrentVersion() int {
	return emk.KeyVersion
}

func (emk *EncryptedMasterKey) GetKeyByVersion(version int) *EncryptedHistoricalKey {
	if version == emk.KeyVersion {
		// Return current key as historical format
		return &EncryptedHistoricalKey{
			KeyVersion: emk.KeyVersion,
			Ciphertext: emk.Ciphertext,
			Nonce:      emk.Nonce,
			Algorithm:  "chacha20poly1305", // or your current algorithm
		}
	}

	for _, key := range emk.PreviousKeys {
		if key.KeyVersion == version {
			return &key
		}
	}
	return nil
}

// KeyEncryptionKey derived from user password
type KeyEncryptionKey struct {
	Key  []byte `json:"key" bson:"key"`
	Salt []byte `json:"salt" bson:"salt"`
}

// PublicKey for asymmetric encryption
type PublicKey struct {
	Key            []byte `json:"key" bson:"key"`
	VerificationID string `json:"verification_id" bson:"verification_id"`
}

// PrivateKey for asymmetric decryption
type PrivateKey struct {
	Key []byte `json:"key" bson:"key"`
}

// EncryptedPrivateKey is the private key encrypted with the master key
type EncryptedPrivateKey struct {
	Ciphertext []byte `json:"ciphertext" bson:"ciphertext"`
	Nonce      []byte `json:"nonce" bson:"nonce"`
}

// RecoveryKey for account recovery
type RecoveryKey struct {
	Key []byte `json:"key" bson:"key"`
}

// EncryptedRecoveryKey is the recovery key encrypted with the master key
type EncryptedRecoveryKey struct {
	Ciphertext []byte `json:"ciphertext" bson:"ciphertext"`
	Nonce      []byte `json:"nonce" bson:"nonce"`
}

// CollectionKey encrypts files in a collection
type CollectionKey struct {
	Key          []byte `json:"key" bson:"key"`
	CollectionID string `json:"collection_id" bson:"collection_id"`
}

// EncryptedCollectionKey is the collection key encrypted with master key
type EncryptedCollectionKey struct {
	Ciphertext   []byte                   `json:"ciphertext" bson:"ciphertext"`
	Nonce        []byte                   `json:"nonce" bson:"nonce"`
	KeyVersion   int                      `json:"key_version" bson:"key_version"`
	RotatedAt    *time.Time               `json:"rotated_at,omitempty" bson:"rotated_at,omitempty"`
	PreviousKeys []EncryptedHistoricalKey `json:"previous_keys,omitempty" bson:"previous_keys,omitempty"`
}

func (eck *EncryptedCollectionKey) NeedsRotation(policy KeyRotationPolicy) bool {
	if eck.RotatedAt == nil {
		return true // Never rotated
	}

	keyAge := time.Since(*eck.RotatedAt)
	return keyAge > policy.MaxKeyAge
}

// UnmarshalJSON custom unmarshaller for EncryptedCollectionKey to handle URL-safe base64 strings.
func (eck *EncryptedCollectionKey) UnmarshalJSON(data []byte) error {
	// Temporary struct to unmarshal into string fields
	type Alias struct {
		Ciphertext string `json:"ciphertext"`
		Nonce      string `json:"nonce"`
	}
	var alias Alias

	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal EncryptedCollectionKey into alias: %w", err)
	}

	// Decode Ciphertext from URL-safe base64
	// libsodium-wrappers to_base64 often uses URLSAFE_NO_PADDING by default or as a common option.
	// base64.RawURLEncoding handles URL-safe base64 without padding.
	if alias.Ciphertext != "" {
		ciphertextBytes, err := base64.RawURLEncoding.DecodeString(alias.Ciphertext)
		if err != nil {
			// Attempt to decode with standard encoding as a fallback if RawURLEncoding fails,
			// though the error "illegal base64 data at input byte 0" for '_' indicates URL-safe.
			ciphertextBytes, err = base64.StdEncoding.DecodeString(alias.Ciphertext)
			if err != nil {
				return fmt.Errorf("failed to decode ciphertext (tried RawURL and Std): %w", err)
			}
		}
		eck.Ciphertext = ciphertextBytes
	}

	// Decode Nonce from URL-safe base64
	if alias.Nonce != "" {
		nonceBytes, err := base64.RawURLEncoding.DecodeString(alias.Nonce)
		if err != nil {
			nonceBytes, err = base64.StdEncoding.DecodeString(alias.Nonce)
			if err != nil {
				return fmt.Errorf("failed to decode nonce (tried RawURL and Std): %w", err)
			}
		}
		eck.Nonce = nonceBytes
	}

	return nil
}

// FileKey encrypts a specific file
type FileKey struct {
	Key    []byte `json:"key" bson:"key"`
	FileID string `json:"file_id" bson:"file_id"`
}

// EncryptedFileKey is the file key encrypted with collection key
type EncryptedFileKey struct {
	Ciphertext   []byte                   `json:"ciphertext" bson:"ciphertext"`
	Nonce        []byte                   `json:"nonce" bson:"nonce"`
	KeyVersion   int                      `json:"key_version" bson:"key_version"`
	RotatedAt    *time.Time               `json:"rotated_at,omitempty" bson:"rotated_at,omitempty"`
	PreviousKeys []EncryptedHistoricalKey `json:"previous_keys,omitempty" bson:"previous_keys,omitempty"`
}

func (eck *EncryptedFileKey) NeedsRotation(policy KeyRotationPolicy) bool {
	if eck.RotatedAt == nil {
		return true // Never rotated
	}

	keyAge := time.Since(*eck.RotatedAt)
	return keyAge > policy.MaxKeyAge
}

// UnmarshalJSON custom unmarshaller for EncryptedFileKey to handle URL-safe base64 strings.
func (efk *EncryptedFileKey) UnmarshalJSON(data []byte) error {
	// Temporary struct to unmarshal into string fields
	type Alias struct {
		Ciphertext string `json:"ciphertext"`
		Nonce      string `json:"nonce"`
	}
	var alias Alias

	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal EncryptedFileKey into alias: %w", err)
	}

	// Decode Ciphertext from URL-safe base64
	if alias.Ciphertext != "" {
		ciphertextBytes, err := base64.RawURLEncoding.DecodeString(alias.Ciphertext)
		if err != nil {
			// Fallback attempt for standard encoding, though URL-safe is expected from client
			ciphertextBytes, err = base64.StdEncoding.DecodeString(alias.Ciphertext)
			if err != nil {
				return fmt.Errorf("failed to decode EncryptedFileKey.Ciphertext (tried RawURL and Std): %w", err)
			}
		}
		efk.Ciphertext = ciphertextBytes
	}

	// Decode Nonce from URL-safe base64
	if alias.Nonce != "" {
		nonceBytes, err := base64.RawURLEncoding.DecodeString(alias.Nonce)
		if err != nil {
			// Fallback attempt for standard encoding
			nonceBytes, err = base64.StdEncoding.DecodeString(alias.Nonce)
			if err != nil {
				return fmt.Errorf("failed to decode EncryptedFileKey.Nonce (tried RawURL and Std): %w", err)
			}
		}
		efk.Nonce = nonceBytes
	}

	return nil
}

// MasterKeyEncryptedWithRecoveryKey allows account recovery
type MasterKeyEncryptedWithRecoveryKey struct {
	Ciphertext []byte `json:"ciphertext" bson:"ciphertext"`
	Nonce      []byte `json:"nonce" bson:"nonce"`
}
