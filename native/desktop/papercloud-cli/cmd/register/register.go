package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/crypto"
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
Use the --skip-remote flag to only save locally without registering with the remote server.

Examples:
		# Register with only required fields
		register --email user@example.com --password mysecret --firstname John --lastname Doe

		# Register with all fields using short flags (note: only some have short flags)
		register -e test@domain.com -p pass123 -f Jane -l Smith -t "America/Toronto" -c "USA" -n "555-1234" --beta-code ABCDE --agree-terms --module 2

		# Register using a mix of short and long flags, enabling all agreements and specifying module
		register --email another@user.net -p anotherpass -f Bob -l Williams --timezone "Europe/London" --agree-terms --agree-promotions --agree-tracking --module 1

		# Save registration information locally without remote registration
		register --email user@example.com --password mysecret --firstname John --lastname Doe --skip-remote`,
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

			// Generate E2EE fields - using our pkg/crypto library
			fmt.Println("Generating secure cryptographic keys...")

			// Generate salt for key derivation
			salt, err := crypto.GenerateRandomBytes(crypto.SaltSize)
			if err != nil {
				fmt.Printf("Error generating salt: %v\n", err)
				return
			}

			// Derive key from password
			params := crypto.DefaultParams()
			keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, salt, params)
			if err != nil {
				fmt.Printf("Error deriving key from password: %v\n", err)
				return
			}

			// Generate master key
			masterKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
			if err != nil {
				fmt.Printf("Error generating master key: %v\n", err)
				return
			}

			// Generate key pair
			publicKey, privateKey, err := crypto.GenerateKeyPair()
			if err != nil {
				fmt.Printf("Error generating key pair: %v\n", err)
				return
			}

			// Generate recovery key
			recoveryKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
			if err != nil {
				fmt.Printf("Error generating recovery key: %v\n", err)
				return
			}

			// Encrypt master key with key encryption key
			encryptedMasterKeyCiphertext, encryptedMasterKeyNonce, err := crypto.EncryptWithSecretBox(masterKey, keyEncryptionKey)
			if err != nil {
				fmt.Printf("Error encrypting master key: %v\n", err)
				return
			}

			// Encrypt private key with master key
			encryptedPrivateKeyCiphertext, encryptedPrivateKeyNonce, err := crypto.EncryptWithSecretBox(privateKey[:], masterKey)
			if err != nil {
				fmt.Printf("Error encrypting private key: %v\n", err)
				return
			}

			// Encrypt recovery key with master key
			encryptedRecoveryKeyCiphertext, encryptedRecoveryKeyNonce, err := crypto.EncryptWithSecretBox(recoveryKey, masterKey)
			if err != nil {
				fmt.Printf("Error encrypting recovery key: %v\n", err)
				return
			}

			// Encrypt master key with recovery key
			masterKeyEncryptedWithRecoveryKeyCiphertext, masterKeyEncryptedWithRecoveryKeyNonce, err := crypto.EncryptWithSecretBox(masterKey, recoveryKey)
			if err != nil {
				fmt.Printf("Error encrypting master key with recovery key: %v\n", err)
				return
			}

			// Create verification ID from public key
			verificationID := crypto.ToBase64(publicKey[:])[:12]

			// Combine nonce and ciphertext for server format
			encryptedMasterKey := crypto.CombineNonceAndCiphertext(encryptedMasterKeyNonce, encryptedMasterKeyCiphertext)
			encryptedPrivateKey := crypto.CombineNonceAndCiphertext(encryptedPrivateKeyNonce, encryptedPrivateKeyCiphertext)
			encryptedRecoveryKey := crypto.CombineNonceAndCiphertext(encryptedRecoveryKeyNonce, encryptedRecoveryKeyCiphertext)
			masterKeyEncryptedWithRecoveryKey := crypto.CombineNonceAndCiphertext(masterKeyEncryptedWithRecoveryKeyNonce, masterKeyEncryptedWithRecoveryKeyCiphertext)

			// Convert to base64 for API
			saltBase64 := crypto.ToBase64(salt)
			publicKeyBase64 := crypto.ToBase64(publicKey[:])
			encryptedMasterKeyBase64 := crypto.ToBase64(encryptedMasterKey)
			encryptedPrivateKeyBase64 := crypto.ToBase64(encryptedPrivateKey)
			encryptedRecoveryKeyBase64 := crypto.ToBase64(encryptedRecoveryKey)
			masterKeyEncryptedWithRecoveryKeyBase64 := crypto.ToBase64(masterKeyEncryptedWithRecoveryKey)

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
				Ciphertext: encryptedMasterKeyCiphertext,
				Nonce:      encryptedMasterKeyNonce,
			}
			if err := configService.SetEncryptedMasterKey(ctx, encryptedMasterKeyObj); err != nil {
				fmt.Printf("Error saving encrypted master key to config: %v\n", err)
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
				var responseData map[string]interface{}
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
