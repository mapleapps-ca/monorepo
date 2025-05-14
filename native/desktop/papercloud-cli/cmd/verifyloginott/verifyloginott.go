// monorepo/native/desktop/papercloud-cli/cmd/verifyloginott/verifyloginott.go
package verifyloginott

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// VerifyOTTRequestPayload represents the data structure sent to the verify-ott endpoint
type VerifyOTTRequestPayload struct {
	Email string `json:"email"`
	OTT   string `json:"ott"`
}

// VerifyOTTResponsePayload represents the response from the verify-ott API
type VerifyOTTResponsePayload struct {
	Salt                string `json:"salt"`
	PublicKey           string `json:"publicKey"`
	EncryptedMasterKey  string `json:"encryptedMasterKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	EncryptedChallenge  string `json:"encryptedChallenge"`
	ChallengeID         string `json:"challengeId"`
}

func VerifyLoginOneTimeTokenUserCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
	var email, ott string

	var cmd = &cobra.Command{
		Use:   "verifyloginott",
		Short: "Verify one-time login token for email",
		Long: `
Verify the OTT and get the encrypted keys and challenge for login.

After requesting a one-time token with requestloginott, use this command to verify it.

Examples:
  # Verify with email and one-time token
  papercloud-cli verifyloginott --email user@example.com --ott 123456
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Verifying one-time login token...")

			if email == "" {
				log.Fatal("Email is required")
			}
			if ott == "" {
				log.Fatal("One-time token (OTT) is required")
			}

			// Sanitize inputs
			email = strings.ToLower(strings.TrimSpace(email))
			ott = strings.TrimSpace(ott)

			ctx := context.Background()

			// Get the server URL from configuration
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				log.Fatalf("Error loading cloud provider address: %v", err)
				return
			}

			// Create the request payload
			verifyPayload := VerifyOTTRequestPayload{
				Email: email,
				OTT:   ott,
			}

			// Convert request to JSON
			jsonData, err := json.Marshal(verifyPayload)
			if err != nil {
				log.Fatalf("Error creating request: %v", err)
				return
			}

			// Make HTTP request to server
			verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-ott", serverURL)
			fmt.Printf("Connecting to: %s\n", verifyURL)

			// Create and execute the HTTP request
			req, err := http.NewRequest("POST", verifyURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatalf("Error creating HTTP request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatalf("Error connecting to server: %v", err)
				return
			}
			defer resp.Body.Close()

			// Read and process the response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Error reading response: %v", err)
				return
			}

			// Check response status code
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				// Try to parse error message if available
				var errorResponse map[string]interface{}
				if err := json.Unmarshal(body, &errorResponse); err == nil {
					if errMsg, ok := errorResponse["message"].(string); ok {
						log.Fatalf("Server error: %s", errMsg)
					} else {
						log.Fatalf("Server returned error status: %s\nResponse body: %s", resp.Status, string(body))
					}
				} else {
					log.Fatalf("Server returned error status: %s\nResponse body: %s", resp.Status, string(body))
				}
				return
			}

			// Parse the response
			var verifyResponse VerifyOTTResponsePayload
			if err := json.Unmarshal(body, &verifyResponse); err != nil {
				log.Fatalf("Error parsing response: %v", err)
				return
			}

			// Retrieve the user by email
			existingUser, err := userRepo.GetByEmail(ctx, email)
			if err != nil {
				log.Fatalf("Error retrieving user: %v", err)
				return
			}

			if existingUser == nil {
				log.Fatalf("User with email %s not found. Please register first.", email)
				return
			}

			// Start a transaction to update the user with the verification data
			if err := userRepo.OpenTransaction(); err != nil {
				log.Fatalf("Error opening transaction: %v", err)
				return
			}

			// Update the user with the verification response data
			updateUserWithVerificationData(existingUser, verifyResponse)

			// Update the user in the repository
			if err := userRepo.UpsertByEmail(ctx, existingUser); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Error updating user with verification data: %v", err)
				return
			}

			// Commit the transaction
			if err := userRepo.CommitTransaction(); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Error committing transaction: %v", err)
				return
			}

			fmt.Println("\n✅ One-time token verified successfully!")
			fmt.Println("You can now proceed to complete the login process with:")
			fmt.Printf("papercloud-cli completelogin --email %s\n", email)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&ott, "ott", "o", "", "One-time token sent by the backend")

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("ott")

	return cmd
}

// updateUserWithVerificationData updates the user model with verification data
func updateUserWithVerificationData(user *user.User, resp VerifyOTTResponsePayload) {
	// Store the verification data temporarily in the user model
	// In a real implementation, we would likely have dedicated fields for these
	// but for now, let's use the available fields or add what's needed

	// Store Encrypted Challenge
	fmt.Printf("DEBUG: Encrypted Challenge from server (b64): %s\n", resp.EncryptedChallenge)
	encryptedChallengeBytes, err := base64.StdEncoding.DecodeString(resp.EncryptedChallenge)
	if err != nil {
		encryptedChallengeBytes, err = base64.RawURLEncoding.DecodeString(resp.EncryptedChallenge)
		if err != nil {
			log.Fatalf("Error decoding encrypted challenge: %v\n", err)
		}
	}
	user.EncryptedChallenge = encryptedChallengeBytes
	fmt.Printf("DEBUG: Encrypted Challenge length after decoding: %d bytes\n", len(encryptedChallengeBytes))

	// Store Salt (decode from base64 if needed)
	salt, err := base64.RawURLEncoding.DecodeString(resp.Salt)
	if err == nil {
		user.PasswordSalt = salt
	}

	// Store Public Key
	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(resp.PublicKey)
	if err == nil {
		user.PublicKey.Key = publicKeyBytes
	}

	// Store Encrypted Master Key
	encMasterKeyBytes, err := base64.RawURLEncoding.DecodeString(resp.EncryptedMasterKey)
	if err == nil && len(encMasterKeyBytes) >= 24 { // Make sure there's enough for nonce and ciphertext
		// Assuming the first 24 bytes are the nonce
		nonceSize := 24 // sodium.crypto_secretbox_NONCEBYTES
		user.EncryptedMasterKey.Nonce = encMasterKeyBytes[:nonceSize]
		user.EncryptedMasterKey.Ciphertext = encMasterKeyBytes[nonceSize:]
	}

	// Store Encrypted Private Key
	encPrivateKeyBytes, err := base64.RawURLEncoding.DecodeString(resp.EncryptedPrivateKey)
	if err == nil && len(encPrivateKeyBytes) >= 24 {
		nonceSize := 24 // sodium.crypto_secretbox_NONCEBYTES
		user.EncryptedPrivateKey.Nonce = encPrivateKeyBytes[:nonceSize]
		user.EncryptedPrivateKey.Ciphertext = encPrivateKeyBytes[nonceSize:]
	}

	// Store ChallengeID - no direct field in User model, so we could use a temporary field
	// For this example, we'll store it in the VerificationID field which seems related
	user.VerificationID = resp.ChallengeID

	// Store the full response in a safe field for use by completelogin
	// This is a simplified approach - in a production system, we'd have a proper preferences store
	// We're using ModifiedAt to update the timestamp
	user.ModifiedAt = time.Now()
}
