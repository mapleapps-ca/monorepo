// monorepo/cloud/mapleapps-backend/internal/iam/domain/keys/kdf.go
package keys

import (
	"fmt"
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/crypto"
)

// KDFParams stores the key derivation function parameters
type KDFParams struct {
	Algorithm   string `json:"algorithm" bson:"algorithm"`     // "argon2id", "pbkdf2", "scrypt"
	Version     string `json:"version" bson:"version"`         // "1.0", "1.1", etc.
	Iterations  uint32 `json:"iterations" bson:"iterations"`   // For PBKDF2 or Argon2 time cost
	Memory      uint32 `json:"memory" bson:"memory"`           // For Argon2 memory in KB
	Parallelism uint8  `json:"parallelism" bson:"parallelism"` // For Argon2 threads
	SaltLength  uint32 `json:"salt_length" bson:"salt_length"` // Salt size in bytes
	KeyLength   uint32 `json:"key_length" bson:"key_length"`   // Output key size in bytes
}

// DefaultKDFParams returns the current recommended KDF parameters
func DefaultKDFParams() KDFParams {
	return KDFParams{
		Algorithm:   crypto.Argon2IDAlgorithm,
		Version:     "1.0",                 // Always starts at 1.0
		Iterations:  crypto.Argon2OpsLimit, // Time cost
		Memory:      crypto.Argon2MemLimit,
		Parallelism: crypto.Argon2Parallelism,
		SaltLength:  crypto.Argon2SaltSize,
		KeyLength:   crypto.Argon2KeySize,
	}
}

// Validate checks if KDF parameters are valid
func (k KDFParams) Validate() error {
	switch k.Algorithm {
	case crypto.Argon2IDAlgorithm:
		if k.Iterations < 1 {
			return fmt.Errorf("argon2id time cost must be >= 1")
		}
		if k.Memory < 1024 {
			return fmt.Errorf("argon2id memory must be >= 1024 KB")
		}
		if k.Parallelism < 1 {
			return fmt.Errorf("argon2id parallelism must be >= 1")
		}
	default:
		return fmt.Errorf("unsupported KDF algorithm: %s", k.Algorithm)
	}

	if k.SaltLength < 8 {
		return fmt.Errorf("salt length must be >= 8 bytes")
	}
	if k.KeyLength < 16 {
		return fmt.Errorf("key length must be >= 16 bytes")
	}

	return nil
}

// KDFUpgradePolicy defines when to upgrade KDF parameters
type KDFUpgradePolicy struct {
	MinimumParams      KDFParams     `json:"minimum_params" bson:"minimum_params"`
	RecommendedParams  KDFParams     `json:"recommended_params" bson:"recommended_params"`
	MaxPasswordAge     time.Duration `json:"max_password_age" bson:"max_password_age"`
	UpgradeOnNextLogin bool          `json:"upgrade_on_next_login" bson:"upgrade_on_next_login"`
	LastUpgradeCheck   time.Time     `json:"last_upgrade_check" bson:"last_upgrade_check"`
}
