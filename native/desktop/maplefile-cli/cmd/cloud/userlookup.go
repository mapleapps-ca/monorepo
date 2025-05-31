// monorepo/native/desktop/maplefile-cli/cmd/healthcheck/healthcheck.go
package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
)

func PublicUserLookupCmd(configService config.ConfigService) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "public-user-lookup",
		Short: "Lookup user",
		Long:  `Command will execute call to backend server to see if a particular email exists and if it does then return public information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing public user lookup...")

			// Validate required fields
			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			// Get the server URL from configuration
			ctx := context.Background()
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				return
			}

			// Make a GET request to the healthcheck endpoint
			publicUserLookupURL := fmt.Sprintf("%s/iam/api/v1/users/lookup?email=%s", serverURL, email)
			fmt.Printf("Connecting to: %s\n", publicUserLookupURL)

			resp, err := http.Get(publicUserLookupURL)
			if err != nil {
				fmt.Printf("Error connecting to server: %v\n", err)
				return
			}
			defer resp.Body.Close()

			// Check if the response was successful
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Server returned error status: %s\n", resp.Status)
				return
			}

			// Read and display the response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			// Parse the JSON response
			var apiResponse struct {
				UserID         string `json:"user_id"`
				Email          string `json:"email"`
				Name           string `json:"name"`       // Optional: for display
				PublicKey      string `json:"public_key"` // Base64 encoded
				VerificationID string `json:"verification_id"`
			}
			if err := json.Unmarshal(body, &apiResponse); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				fmt.Printf("Raw response: %s\n", string(body))
				return
			}

			// Display the status
			fmt.Printf("User ID: %s\n", apiResponse.UserID)
			fmt.Printf("Email: %s\n", apiResponse.Email)
			fmt.Printf("Name: %s\n", apiResponse.Name)
			fmt.Printf("PublicKey (Base64 encoded): %s\n", apiResponse.PublicKey)
			fmt.Printf("VerificationID: %s\n", apiResponse.VerificationID)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
