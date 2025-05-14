// monorepo/native/desktop/papercloud-cli/cmd/requestloginott/requestloginott.go
package requestloginott

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

// RequestOTTPayload represents the data structure sent to the request-ott endpoint
type RequestOTTPayload struct {
	Email string `json:"email"`
}

func RequestLoginOneTimeTokenUserCmd(configService config.ConfigService) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "requestloginott",
		Short: "Request one-time login token for email",
		Long: `
Command will execute login command and user will get credentials to make API calls to their account.

After registration and email verification, use this command to log in to your account.

Examples:
  # Request login OTT with email
  papercloud-cli requestloginott --email user@example.com
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Requesting login one-time token...")

			if email == "" {
				log.Fatal("Email is required")
			}

			// Sanitize inputs
			email = strings.ToLower(strings.TrimSpace(email))

			// Get the server URL from configuration
			ctx := context.Background()
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				log.Fatalf("Error loading cloud provider address: %v", err)
				return
			}

			// Create the request payload
			requestPayload := RequestOTTPayload{
				Email: email,
			}

			// Convert request to JSON
			jsonData, err := json.Marshal(requestPayload)
			if err != nil {
				log.Fatalf("Error creating request: %v", err)
				return
			}

			// Make HTTP request to server
			requestURL := fmt.Sprintf("%s/iam/api/v1/request-ott", serverURL)
			fmt.Printf("Connecting to: %s\n", requestURL)

			// Create and execute the HTTP request
			req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
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
				var errorResponse map[string]any
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

			fmt.Println("\n✅ One-time login token request successful!")
			fmt.Println("Please check your email for a one-time token.")
			fmt.Println("\nOnce you receive the token, run the following command to verify it:")
			fmt.Printf("papercloud-cli verifyloginott --email %s --ott YOUR_TOKEN\n", email)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
