package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	sstring "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securestring"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type PasswordProvider interface {
	GenerateHashFromPassword(password *sstring.SecureString) (string, error)
	ComparePasswordAndHash(password *sstring.SecureString, hash string) (bool, error)
	AlgorithmName() string
	GenerateSecureRandomBytes(length int) ([]byte, error)
	GenerateSecureRandomString(length int) (string, error)
}

type passwordProvider struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewPasswordProvider() PasswordProvider {
	// DEVELOPERS NOTE:
	// The following code was copy and pasted from: "How to Hash and Verify Passwords With Argon2 in Go" via https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

	// Establish the parameters to use for Argon2.
	return &passwordProvider{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

// GenerateHashFromPassword function takes the plaintext string and returns an Argon2 hashed string.
func (p *passwordProvider) GenerateHashFromPassword(password *sstring.SecureString) (string, error) {
	fmt.Println("GenerateHashFromPassword: Starting")

	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		fmt.Printf("GenerateHashFromPassword: Failed to generate salt: %v\n", err)
		return "", err
	}
	fmt.Println("GenerateHashFromPassword: Salt generated")

	passwordBytes := password.Bytes()
	fmt.Printf("GenerateHashFromPassword: Getting password bytes, length: %d\n", len(passwordBytes))

	fmt.Println("GenerateHashFromPassword: Calling argon2.IDKey...")
	hash := argon2.IDKey(passwordBytes, salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	fmt.Println("GenerateHashFromPassword: argon2.IDKey completed")

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	fmt.Println("GenerateHashFromPassword: Completed")
	return encodedHash, nil
}

// CheckPasswordHash function checks the plaintext string and hash string and returns either true
// or false depending.
func (p *passwordProvider) ComparePasswordAndHash(password *sstring.SecureString, encodedHash string) (match bool, err error) {
	// DEVELOPERS NOTE:
	// The following code was copy and pasted from: "How to Hash and Verify Passwords With Argon2 in Go" via https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey(password.Bytes(), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// AlgorithmName function returns the algorithm used for hashing.
func (p *passwordProvider) AlgorithmName() string {
	return "argon2id"
}

func generateRandomBytes(n uint32) ([]byte, error) {
	// DEVELOPERS NOTE:
	// The following code was copy and pasted from: "How to Hash and Verify Passwords With Argon2 in Go" via https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func decodeHash(encodedHash string) (p *passwordProvider, salt, hash []byte, err error) {
	// DEVELOPERS NOTE:
	// The following code was copy and pasted from: "How to Hash and Verify Passwords With Argon2 in Go" via https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &passwordProvider{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

// GenerateSecureRandomBytes generates a secure random byte slice of the specified length.
func (p *passwordProvider) GenerateSecureRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure random bytes: %v", err)
	}
	return bytes, nil
}

// GenerateSecureRandomString generates a secure random string of the specified length.
func (p *passwordProvider) GenerateSecureRandomString(length int) (string, error) {
	bytes, err := p.GenerateSecureRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
