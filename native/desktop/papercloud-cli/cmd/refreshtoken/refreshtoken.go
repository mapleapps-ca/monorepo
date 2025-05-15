// monorepo/native/desktop/papercloud-cli/cmd/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// RefreshTokenRequest represents the data structure sent to the token refresh endpoint
type RefreshTokenRequest struct {
	Value string `json:"value"`
}

// RefreshTokenResponse represents the response from the token refresh API
type RefreshTokenResponse struct {
	Email                  string    `json:"username"`
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
}

func RefreshTokenCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "refreshtoken",
		Short: "Force refresh current authenticated user's token",
		Long: `
Forces a refresh of the current authenticated user's access token by using
the refresh token. This is useful when the access token has expired but
the refresh token is still valid.

This command does not take any arguments. It uses the refresh token
stored locally for the currently authenticated user.

Example:
  papercloud-cli refreshtoken
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Refreshing authentication token...")

			ctx := context.Background()

			// Get the current user's email from configuration
			email, err := configService.GetEmail(ctx)
			if err != nil {
				log.Fatalf("Error getting authenticated user email: %v", err)
				return
			}

			if email == "" {
				log.Fatal("No authenticated user found. Please login first.")
				return
			}

			// Get the user details from the repository
			userData, err := userRepo.GetByEmail(ctx, email)
			if err != nil {
				log.Fatalf("Error retrieving user data: %v", err)
				return
			}

			if userData == nil {
				log.Fatalf("User with email %s not found", email)
				return
			}

			// Check if the user has a refresh token
			if userData.RefreshToken == "" {
				log.Fatal("No refresh token found. Please login again.")
				return
			}

			// Check if the refresh token has expired
			if time.Now().After(userData.RefreshTokenExpiryTime) {
				log.Fatal("Refresh token has expired. Please login again.")
				return
			}

			// Get the server URL from configuration
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				log.Fatalf("Error loading cloud provider address: %v", err)
				return
			}

			// Create the request payload
			refreshReq := RefreshTokenRequest{
				Value: userData.RefreshToken,
			}

			// Convert request to JSON
			jsonData, err := json.Marshal(refreshReq)
			if err != nil {
				log.Fatalf("Error creating request: %v", err)
				return
			}

			// Make HTTP request to server
			refreshURL := fmt.Sprintf("%s/iam/api/v1/token/refresh", serverURL)
			fmt.Printf("Connecting to: %s\n", refreshURL)

			// Create and execute the HTTP request
			req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer(jsonData))
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

			// Parse the response
			var tokenResponse RefreshTokenResponse
			if err := json.Unmarshal(body, &tokenResponse); err != nil {
				log.Fatalf("Error parsing response: %v", err)
				return
			}

			// Start a transaction to update the user with the new tokens
			if err := userRepo.OpenTransaction(); err != nil {
				log.Fatalf("Error opening transaction: %v", err)
				return
			}

			// Update user with the new tokens
			userData.AccessToken = tokenResponse.AccessToken
			userData.AccessTokenExpiryTime = tokenResponse.AccessTokenExpiryDate
			userData.RefreshToken = tokenResponse.RefreshToken
			userData.RefreshTokenExpiryTime = tokenResponse.RefreshTokenExpiryDate
			userData.ModifiedAt = time.Now()

			// Update the user in the repository
			if err := userRepo.UpsertByEmail(ctx, userData); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Error updating user with new tokens: %v", err)
				return
			}

			// Commit the transaction
			if err := userRepo.CommitTransaction(); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Error committing transaction: %v", err)
				return
			}

			fmt.Println("\n✅ Authentication tokens refreshed successfully!")
			fmt.Printf("Access Token expires: %s\n", tokenResponse.AccessTokenExpiryDate.Format(time.RFC3339))
			fmt.Printf("Refresh Token expires: %s\n", tokenResponse.RefreshTokenExpiryDate.Format(time.RFC3339))
		},
	}

	return cmd
}
