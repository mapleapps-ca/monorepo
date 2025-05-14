// monorepo/native/desktop/papercloud-cli/internal/cmd/register/register.go
package register

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/keys"
)

// Constants matching the backend but with reduced memory for Argon2
const (
	// Key sizes
	MasterKeySize        = 32 // 256-bit
	KeyEncryptionKeySize = 32
	CollectionKeySize    = 32
	FileKeySize          = 32
	RecoveryKeySize      = 32

	// Sodium/NaCl constants
	NonceSize         = 24
	PublicKeySize     = 32
	PrivateKeySize    = 32
	SealedBoxOverhead = 16

	// Argon2 parameters - reduced for CLI usage
	Argon2MemLimit    = 4 * 1024 * 1024 // 4 MB instead of 16 MB
	Argon2OpsLimit    = 1               // 1 iteration instead of 3
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16
)

// RegisterRequest represents the data structure needed for user registration
type RegisterRequest struct {
	// Personal information
	BetaAccessCode                                 string `json:"beta_access_code"`
	FirstName                                      string `json:"first_name"`
	LastName                                       string `json:"last_name"`
	Email                                          string `json:"email"`
	Phone                                          string `json:"phone,omitempty"`
	Country                                        string `json:"country,omitempty"`
	CountryOther                                   string `json:"country_other,omitempty"`
	Timezone                                       string `json:"timezone"`
	AgreeTermsOfService                            bool   `json:"agree_terms_of_service"`
	AgreePromotions                                bool   `json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
	Module                                         int    `json:"module"`

	// E2EE related fields
	Salt                              string `json:"salt"`
	PublicKey                         string `json:"publicKey"`
	EncryptedMasterKey                string `json:"encryptedMasterKey"`
	EncryptedPrivateKey               string `json:"encryptedPrivateKey"`
	EncryptedRecoveryKey              string `json:"encryptedRecoveryKey"`
	MasterKeyEncryptedWithRecoveryKey string `json:"masterKeyEncryptedWithRecoveryKey"`
	VerificationID                    string `json:"verificationID"`
}

// Crypto utility functions needed for registration

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// DeriveKeyFromPassword derives a key from a password using Argon2id
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != Argon2SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d, got %d", Argon2SaltSize, len(salt))
	}

	// Use modified parameters for CLI use
	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2OpsLimit,
		Argon2MemLimit,
		Argon2Parallelism,
		Argon2KeySize,
	)

	return key, nil
}

// EncryptData represents encrypted data with its nonce
type EncryptData struct {
	Ciphertext []byte
	Nonce      []byte
}

// EncryptWithSecretBox encrypts data with a symmetric key
func EncryptWithSecretBox(data, key []byte) (*EncryptData, error) {
	if len(key) != MasterKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", MasterKeySize, len(key))
	}

	// Generate nonce
	nonce, err := GenerateRandomBytes(NonceSize)
	if err != nil {
		return nil, err
	}

	// Create a fixed-size array from slice for secretbox
	var keyArray [32]byte
	copy(keyArray[:], key)

	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt
	ciphertext := secretbox.Seal(nil, data, &nonceArray, &keyArray)

	return &EncryptData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// The EncryptWithBoxSeal function that correctly implements crypto_box_seal
func EncryptWithBoxSeal(message []byte, recipientPK []byte) ([]byte, error) {
	if len(recipientPK) != PublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", PublicKeySize)
	}

	// Create a fixed-size array for the recipient's public key
	var recipientPKArray [32]byte
	copy(recipientPKArray[:], recipientPK)

	// Generate an ephemeral keypair
	ephemeralPK, ephemeralSK, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	nonce, err := GenerateRandomBytes(NonceSize)
	if err != nil {
		return nil, err
	}
	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt the message
	ciphertext := box.Seal(nil, message, &nonceArray, &recipientPKArray, ephemeralSK)

	// Result format: ephemeral_public_key || nonce || ciphertext
	result := make([]byte, PublicKeySize+NonceSize+len(ciphertext))
	copy(result[:PublicKeySize], ephemeralPK[:])
	copy(result[PublicKeySize:PublicKeySize+NonceSize], nonce)
	copy(result[PublicKeySize+NonceSize:], ciphertext)

	return result, nil
}

func RegisterCmd(configService config.ConfigService) *cobra.Command {
	var email, password, firstName, lastName, timezone, country, phone, betaAccessCode string
	var agreeTerms, agreePromotions, agreeTracking, skipRemoteRegistration bool
	var module int

	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Register an account on PaperCloud",
		Long: `Register a new user account in the system.

This command requires you to provide an email, password, first name, and last name.
You can optionally provide timezone, country, phone number, a beta access code,
specify agreement to terms, promotions, and tracking, and specify the registration module.

Registration information will be saved locally before being sent to the remote server.
Use the --skip-remote flag to only save locally without registering with the remote server.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if email == "" || password == "" || firstName == "" || lastName == "" {
				fmt.Println("Error: email, password, first name, and last name are required")
				return
			}

			if !agreeTerms {
				fmt.Println("Error: you must agree to the Terms of Service")
				return
			}

			// Set default module value if not specified
			if module <= 0 {
				module = 1 // Default to module 1 (PaperCloud)
			}

			// Generate E2EE fields
			fmt.Println("Generating secure cryptographic keys...")

			// Generate salt for key derivation
			salt, err := GenerateRandomBytes(Argon2SaltSize)
			if err != nil {
				fmt.Printf("Error generating salt: %v\n", err)
				return
			}

			// Derive key from password
			fmt.Println("Deriving key from password...")
			keyEncryptionKey, err := DeriveKeyFromPassword(password, salt)
			if err != nil {
				fmt.Printf("Error deriving key from password: %v\n", err)
				return
			}

			// Generate master key
			fmt.Println("Generating master key...")
			masterKey, err := GenerateRandomBytes(MasterKeySize)
			if err != nil {
				fmt.Printf("Error generating master key: %v\n", err)
				return
			}

			// Generate key pair
			fmt.Println("Generating key pair...")
			pubKey, privKey, err := box.GenerateKey(rand.Reader)
			if err != nil {
				fmt.Printf("Error generating key pair: %v\n", err)
				return
			}
			publicKey := pubKey[:]
			privateKey := privKey[:]

			// Generate recovery key
			fmt.Println("Generating recovery key...")
			recoveryKey, err := GenerateRandomBytes(RecoveryKeySize)
			if err != nil {
				fmt.Printf("Error generating recovery key: %v\n", err)
				return
			}

			// Encrypt master key with key encryption key
			fmt.Println("Encrypting master key...")
			encryptedMasterKey, err := EncryptWithSecretBox(masterKey, keyEncryptionKey)
			if err != nil {
				fmt.Printf("Error encrypting master key: %v\n", err)
				return
			}

			// Encrypt private key with master key
			fmt.Println("Encrypting private key...")
			encryptedPrivateKey, err := EncryptWithSecretBox(privateKey, masterKey)
			if err != nil {
				fmt.Printf("Error encrypting private key: %v\n", err)
				return
			}

			// Encrypt recovery key with master key
			fmt.Println("Encrypting recovery key...")
			encryptedRecoveryKey, err := EncryptWithSecretBox(recoveryKey, masterKey)
			if err != nil {
				fmt.Printf("Error encrypting recovery key: %v\n", err)
				return
			}

			// Encrypt master key with recovery key
			fmt.Println("Encrypting master key with recovery key...")
			masterKeyEncryptedWithRecoveryKey, err := EncryptWithSecretBox(masterKey, recoveryKey)
			if err != nil {
				fmt.Printf("Error encrypting master key with recovery key: %v\n", err)
				return
			}

			// Create verification ID from public key - simple approach for now
			fmt.Println("Generating verification ID...")
			verificationID := base64.URLEncoding.EncodeToString(publicKey)[:12]

			// Combine nonce and ciphertext for each encrypted value
			encryptedMasterKeyBytes := append(encryptedMasterKey.Nonce, encryptedMasterKey.Ciphertext...)
			encryptedPrivateKeyBytes := append(encryptedPrivateKey.Nonce, encryptedPrivateKey.Ciphertext...)
			encryptedRecoveryKeyBytes := append(encryptedRecoveryKey.Nonce, encryptedRecoveryKey.Ciphertext...)
			masterKeyEncryptedWithRecoveryKeyBytes := append(masterKeyEncryptedWithRecoveryKey.Nonce, masterKeyEncryptedWithRecoveryKey.Ciphertext...)

			// Convert to base64 for API
			saltBase64 := base64.RawURLEncoding.EncodeToString(salt)
			publicKeyBase64 := base64.RawURLEncoding.EncodeToString(publicKey)
			encryptedMasterKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedMasterKeyBytes)
			encryptedPrivateKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedPrivateKeyBytes)
			encryptedRecoveryKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedRecoveryKeyBytes)
			masterKeyEncryptedWithRecoveryKeyBase64 := base64.RawURLEncoding.EncodeToString(masterKeyEncryptedWithRecoveryKeyBytes)

			// Save information to local config
			fmt.Println("Saving registration information locally...")

			// Save email to config
			if err := configService.SetEmail(ctx, email); err != nil {
				fmt.Printf("Error saving email to config: %v\n", err)
				return
			}

			// Save password salt to config
			if err := configService.Set(ctx, "password_salt", salt); err != nil {
				fmt.Printf("Error saving password salt to config: %v\n", err)
				return
			}

			// Save verification ID to config
			if err := configService.Set(ctx, "verification_id", verificationID); err != nil {
				fmt.Printf("Error saving verification ID to config: %v\n", err)
				return
			}

			// Save encrypted master key
			encryptedMasterKeyObj := keys.EncryptedMasterKey{
				Ciphertext: encryptedMasterKey.Ciphertext,
				Nonce:      encryptedMasterKey.Nonce,
			}
			if err := configService.SetEncryptedMasterKey(ctx, encryptedMasterKeyObj); err != nil {
				fmt.Printf("Error saving encrypted master key to config: %v\n", err)
				return
			}

			// Save public key
			publicKeyObj := keys.PublicKey{
				Key:            publicKey,
				VerificationID: verificationID,
			}
			if err := configService.SetPublicKey(ctx, publicKeyObj); err != nil {
				fmt.Printf("Error saving public key to config: %v\n", err)
				return
			}

			// Save encrypted private key
			encryptedPrivateKeyObj := keys.EncryptedPrivateKey{
				Ciphertext: encryptedPrivateKey.Ciphertext,
				Nonce:      encryptedPrivateKey.Nonce,
			}
			if err := configService.SetEncryptedPrivateKey(ctx, encryptedPrivateKeyObj); err != nil {
				fmt.Printf("Error saving encrypted private key to config: %v\n", err)
				return
			}

			// Save encrypted recovery key
			encryptedRecoveryKeyObj := keys.EncryptedRecoveryKey{
				Ciphertext: encryptedRecoveryKey.Ciphertext,
				Nonce:      encryptedRecoveryKey.Nonce,
			}
			if err := configService.SetEncryptedRecoveryKey(ctx, encryptedRecoveryKeyObj); err != nil {
				fmt.Printf("Error saving encrypted recovery key to config: %v\n", err)
				return
			}

			// Save master key encrypted with recovery key
			masterKeyEncryptedWithRecoveryKeyObj := keys.MasterKeyEncryptedWithRecoveryKey{
				Ciphertext: masterKeyEncryptedWithRecoveryKey.Ciphertext,
				Nonce:      masterKeyEncryptedWithRecoveryKey.Nonce,
			}
			if err := configService.SetMasterKeyEncryptedWithRecoveryKey(ctx, masterKeyEncryptedWithRecoveryKeyObj); err != nil {
				fmt.Printf("Error saving master key encrypted with recovery key to config: %v\n", err)
				return
			}

			// Save other user info
			if err := configService.Set(ctx, "first_name", firstName); err != nil {
				fmt.Printf("Error saving first name to config: %v\n", err)
				return
			}

			if err := configService.Set(ctx, "last_name", lastName); err != nil {
				fmt.Printf("Error saving last name to config: %v\n", err)
				return
			}

			// Save timezone
			if err := configService.Set(ctx, "timezone", timezone); err != nil {
				fmt.Printf("Error saving timezone to config: %v\n", err)
				return
			}

			// Only proceed with remote registration if not skipped
			if !skipRemoteRegistration {
				// Get the server URL from configuration
				serverURL, err := configService.GetCloudProviderAddress(ctx)
				if err != nil {
					fmt.Printf("Error loading cloud provider address: %v\n", err)
					return
				}

				// Create registration request
				registerReq := RegisterRequest{
					Email:               strings.ToLower(email),
					FirstName:           firstName,
					LastName:            lastName,
					Phone:               phone,
					Country:             country,
					Timezone:            timezone,
					BetaAccessCode:      betaAccessCode,
					Module:              module,
					AgreeTermsOfService: agreeTerms,
					AgreePromotions:     agreePromotions,
					AgreeToTrackingAcrossThirdPartyAppsAndServices: agreeTracking,

					// Add E2EE fields
					Salt:                              saltBase64,
					PublicKey:                         publicKeyBase64,
					EncryptedMasterKey:                encryptedMasterKeyBase64,
					EncryptedPrivateKey:               encryptedPrivateKeyBase64,
					EncryptedRecoveryKey:              encryptedRecoveryKeyBase64,
					MasterKeyEncryptedWithRecoveryKey: masterKeyEncryptedWithRecoveryKeyBase64,
					VerificationID:                    verificationID,
				}

				// Convert request to JSON
				jsonData, err := json.Marshal(registerReq)
				if err != nil {
					fmt.Printf("Error creating request: %v\n", err)
					return
				}

				// Make HTTP request to server
				fmt.Println("Sending registration request to remote server...")
				registerURL := fmt.Sprintf("%s/iam/api/v1/register", serverURL)

				// Create and execute the HTTP request
				req, err := http.NewRequest("POST", registerURL, bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("Error creating HTTP request: %v\n", err)
					return
				}

				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Error connecting to server: %v\n", err)
					return
				}
				defer resp.Body.Close()

				// Read and process the response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response: %v\n", err)
					return
				}

				// Check response status code
				if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
					fmt.Printf("Server returned error status: %s\n", resp.Status)
					fmt.Printf("Response body: %s\n", string(body))
					return
				}

				// Parse and display the response
				var responseData map[string]any
				if err := json.Unmarshal(body, &responseData); err != nil {
					fmt.Printf("Error parsing response: %v\n", err)
					fmt.Printf("Raw response: %s\n", string(body))
					return
				}

				// Display success message
				fmt.Println("\n✅ Registration successful!")
				fmt.Println("Please check your email for verification instructions.")
				fmt.Println("\nIMPORTANT: Please ensure you have saved your password securely.")
				fmt.Println("You will need it to log in to your account.")
			} else {
				fmt.Println("\n✅ Registration information saved locally.")
				fmt.Println("To complete registration with the remote server, run the command again without the --skip-remote flag.")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (required)")
	cmd.Flags().StringVarP(&firstName, "firstname", "f", "", "First name for the user (required)")
	cmd.Flags().StringVarP(&lastName, "lastname", "l", "", "Last name for the user (required)")
	cmd.Flags().StringVarP(&timezone, "timezone", "t", "UTC", "Timezone for the user (e.g., \"America/New_York\")")
	cmd.Flags().StringVarP(&country, "country", "c", "Canada", "Country for the user")
	cmd.Flags().StringVarP(&phone, "phone", "n", "", "Phone number for the user")
	cmd.Flags().StringVar(&betaAccessCode, "beta-code", "", "Beta access code (if required)")
	cmd.Flags().BoolVar(&agreeTerms, "agree-terms", false, "Agree to Terms of Service")
	cmd.Flags().BoolVar(&agreePromotions, "agree-promotions", false, "Agree to receive promotions")
	cmd.Flags().BoolVar(&agreeTracking, "agree-tracking", false, "Agree to tracking across third-party apps and services")
	cmd.Flags().IntVarP(&module, "module", "m", 0, "Module the user is registering for")
	cmd.Flags().BoolVar(&skipRemoteRegistration, "skip-remote", false, "Skip remote registration and only save locally")

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("firstname")
	cmd.MarkFlagRequired("lastname")

	return cmd
}
