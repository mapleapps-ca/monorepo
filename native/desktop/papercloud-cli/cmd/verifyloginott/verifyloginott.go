// monorepo/native/desktop/papercloud-cli/cmd/verifyloginott/verifyloginott.go
package verifyloginott

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	var debugMode bool

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

			if debugMode {
				fmt.Printf("DEBUG: Sending request to %s/iam/api/v1/verify-ott\n", serverURL)
				fmt.Printf("DEBUG: Request payload: %s\n", string(jsonData))
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

			if debugMode {
				fmt.Printf("DEBUG: Server response code: %d\n", resp.StatusCode)
				fmt.Printf("DEBUG: Response body: %s\n", string(body))
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

			if debugMode {
				fmt.Println("DEBUG: Successfully parsed verify response")
				fmt.Printf("DEBUG: Challenge ID: %s\n", verifyResponse.ChallengeID)
				fmt.Printf("DEBUG: Salt length: %d\n", len(verifyResponse.Salt))
				fmt.Printf("DEBUG: Encrypted Challenge length: %d\n", len(verifyResponse.EncryptedChallenge))
			}

			// Create data directory if it doesn't exist
			appDirPath, err := configService.GetAppDirPath(ctx)
			if err != nil {
				log.Fatalf("Error getting app directory path: %v", err)
				return
			}

			dataDir := filepath.Join(appDirPath, "auth_data")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				log.Fatalf("Failed to create data directory: %v", err)
				return
			}

			// Save the verification data to a file that completelogin can access
			// Using email as part of the filename for user-specific data
			emailHash := hashEmail(email)
			dataFile := filepath.Join(dataDir, fmt.Sprintf("verify_data_%s.json", emailHash))

			dataJSON, err := json.Marshal(verifyResponse)
			if err != nil {
				log.Fatalf("Failed to serialize verification data: %v", err)
				return
			}

			if err := os.WriteFile(dataFile, dataJSON, 0600); err != nil {
				log.Fatalf("Failed to save verification data: %v", err)
				return
			}

			fmt.Println("\n✅ Login verification successful!")
			fmt.Println("You can now proceed to complete the login process with:")
			fmt.Printf("papercloud-cli completelogin --email %s --password YOUR_PASSWORD\n", email)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&ott, "ott", "o", "", "One-time token sent by the backend")
	cmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug output")

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("ott")

	return cmd
}

// Simple helper function to create a deterministic string from email for filename purposes
func hashEmail(email string) string {
	// This is a simple approach for demonstration purposes
	// In a real app, consider using a proper hashing function
	email = strings.ToLower(strings.TrimSpace(email))
	hash := 0
	for i := 0; i < len(email); i++ {
		hash = hash*31 + int(email[i])
	}
	return fmt.Sprintf("%x", hash)
}
