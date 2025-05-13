// monorepo/native/desktop/papercloud-cli/cmd/remote/register.go
package remote

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/spf13/cobra"
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

// generateRandomBytes creates cryptographically secure random bytes
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// generateDummyE2EEFields generates placeholder values for E2EE fields
// In a real implementation, these would be properly generated with actual encryption
func generateDummyE2EEFields() (map[string]string, error) {
	result := make(map[string]string)

	// Generate a verification ID (using UUID format)
	verificationID := fmt.Sprintf("%x-%x-%x-%x-%x",
		time.Now().UnixNano()&0xffffffff,
		time.Now().UnixNano()>>32&0xffff,
		time.Now().UnixNano()>>48&0xffff,
		time.Now().UnixNano()&0xffff,
		time.Now().UnixNano()>>16&0xffffffffffff)

	// Generate random bytes for various fields
	salt, err := generateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	masterKey, err := generateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	privateKey, err := generateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	recoveryKey, err := generateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	// Create SHA-256 hashes and encode them as base64 to simulate encryption
	masterKeyHash := sha256.Sum256(masterKey)
	privateKeyHash := sha256.Sum256(privateKey)
	recoveryKeyHash := sha256.Sum256(recoveryKey)
	masterWithRecoveryHash := sha256.Sum256(append(masterKey, recoveryKey...))

	// Store the encoded values
	result["salt"] = base64.RawURLEncoding.EncodeToString(salt)
	result["publicKey"] = base64.RawURLEncoding.EncodeToString(privateKeyHash[:])
	result["encryptedMasterKey"] = base64.RawURLEncoding.EncodeToString(masterKeyHash[:])
	result["encryptedPrivateKey"] = base64.RawURLEncoding.EncodeToString(privateKeyHash[:])
	result["encryptedRecoveryKey"] = base64.RawURLEncoding.EncodeToString(recoveryKeyHash[:])
	result["masterKeyEncryptedWithRecoveryKey"] = base64.RawURLEncoding.EncodeToString(masterWithRecoveryHash[:])
	result["verificationID"] = verificationID

	return result, nil
}

func RegisterUserCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	var email, password, firstName, lastName, timezone, country, phone, betaAccessCode string
	var agreeTerms, agreePromotions, agreeTracking bool
	var module int

	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Register user account",
		Long: `Register a new user account in the system.

This command requires you to provide an email, password, first name, and last name.
You can optionally provide timezone, country, phone number, a beta access code,
specify agreement to terms, promotions, and tracking, and specify the registration module.

Examples:
		# Register with only required fields
		register --email user@example.com --password mysecret --firstname John --lastname Doe

		# Register with all fields using short flags (note: only some have short flags)
		register -e test@domain.com -p pass123 -f Jane -l Smith -t "America/Toronto" -c "USA" -n "555-1234" --beta-code ABCDE --agree-terms --module 2

		# Register using a mix of short and long flags, enabling all agreements and specifying module
		register --email another@user.net -p anotherpass -f Bob -l Williams --timezone "Europe/London" --agree-terms --agree-promotions --agree-tracking --module 1`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get the server URL from configuration
			ctx := cmd.Context()
			serverURL, err := configUseCase.GetCloudProviderAddress(ctx)
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				return
			}

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
			e2eeFields, err := generateDummyE2EEFields()
			if err != nil {
				fmt.Printf("Error generating encryption fields: %v\n", err)
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
				Salt:                              e2eeFields["salt"],
				PublicKey:                         e2eeFields["publicKey"],
				EncryptedMasterKey:                e2eeFields["encryptedMasterKey"],
				EncryptedPrivateKey:               e2eeFields["encryptedPrivateKey"],
				EncryptedRecoveryKey:              e2eeFields["encryptedRecoveryKey"],
				MasterKeyEncryptedWithRecoveryKey: e2eeFields["masterKeyEncryptedWithRecoveryKey"],
				VerificationID:                    e2eeFields["verificationID"],
			}

			// Convert request to JSON
			jsonData, err := json.Marshal(registerReq)
			if err != nil {
				fmt.Printf("Error creating request: %v\n", err)
				return
			}

			// Make HTTP request to server
			fmt.Println("Sending registration request to server...")
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

			// Store the recovery key (in a real application, this would be displayed to the user)
			// fmt.Printf("\nYour recovery key: %s\n", recoveryKey)
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

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("firstname")
	cmd.MarkFlagRequired("lastname")

	return cmd
}
