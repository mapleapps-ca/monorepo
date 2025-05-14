// monorepo/native/desktop/papercloud-cli/cmd/verifyemail/verifyemail.go
package verifyemail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// VerifyEmailRequest represents the data needed to verify an email
type VerifyEmailRequest struct {
	Code string `json:"code"`
}

// VerifyEmailResponse represents the expected response from the server
type VerifyEmailResponse struct {
	Message  string `json:"message"`
	UserRole int    `json:"user_role"`
	Status   int    `json:"profile_verification_status,omitempty"`
}

func VerifyEmailCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
	var verificationCode string

	var cmd = &cobra.Command{
		Use:   "verify-email",
		Short: "Verify your email address",
		Long: `Verify your email address by submitting the verification code you received by email.

After registering, you should receive an email with a verification code.
Use this command to verify your email address and activate your account.

Example:
  # Verify email with a code provided as an argument
  verify-email 123456

  # Verify email with a code provided as a flag
  verify-email --code 123456`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Check if code was provided as an argument
			if len(args) > 0 && verificationCode == "" {
				verificationCode = args[0]
			}

			// Validate verification code
			if verificationCode == "" {
				fmt.Println("Error: Verification code is required")
				fmt.Println("Please provide the code you received by email after registration.")
				fmt.Println("Usage: verify-email --code YOUR_CODE")
				return
			}

			// List all users to find a registered one
			users, err := userRepo.ListAll(ctx)
			if err != nil || len(users) == 0 {
				fmt.Println("Error: No registered user found.")
				fmt.Println("Please register first using the 'register' command.")
				return
			}

			// Get the first user - this is simple but in a real app
			// we might want to allow the user to select which account to verify
			currentUser := users[0]

			// Get the server URL from configuration
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				fmt.Printf("Error loading configuration: %v\n", err)
				return
			}

			// Create the request payload
			verifyReq := VerifyEmailRequest{
				Code: verificationCode,
			}

			// Convert request to JSON
			jsonData, err := json.Marshal(verifyReq)
			if err != nil {
				fmt.Printf("Error creating request: %v\n", err)
				return
			}

			// Make HTTP request to server
			fmt.Println("Verifying email...")
			verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-email-code", serverURL)

			// Create and execute the HTTP request
			req, err := http.NewRequest("POST", verifyURL, bytes.NewBuffer(jsonData))
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

				// Try to parse error message from response
				var errorResponse map[string]interface{}
				if err := json.Unmarshal(body, &errorResponse); err == nil {
					if errMsg, ok := errorResponse["message"].(string); ok {
						fmt.Printf("Error: %s\n", errMsg)
					} else if errField, ok := errorResponse["code"].(string); ok {
						fmt.Printf("Error: %s\n", errField)
					} else {
						fmt.Printf("Response body: %s\n", string(body))
					}
				} else {
					fmt.Printf("Response body: %s\n", string(body))
				}
				return
			}

			// Parse and display the success response
			var responseData VerifyEmailResponse
			if err := json.Unmarshal(body, &responseData); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				fmt.Printf("Raw response: %s\n", string(body))
				return
			}

			// Update user with verified status
			if err := userRepo.OpenTransaction(); err != nil {
				fmt.Printf("Error opening transaction: %v\n", err)
				return
			}

			// Update user's verification status
			currentUser.WasEmailVerified = true
			currentUser.Role = int8(responseData.UserRole)
			currentUser.Status = user.UserStatusActive
			currentUser.ModifiedAt = time.Now()

			// Save updated user
			if err := userRepo.UpsertByEmail(ctx, currentUser); err != nil {
				fmt.Printf("Error updating user verification status: %v\n", err)
				userRepo.DiscardTransaction()
				return
			}

			// Commit transaction
			if err := userRepo.CommitTransaction(); err != nil {
				fmt.Printf("Error committing transaction: %v\n", err)
				userRepo.DiscardTransaction()
				return
			}

			// Display success message
			fmt.Println("\n✅ Email verification successful!")
			fmt.Printf("\n%s\n", responseData.Message)

			// Display user role information
			switch responseData.UserRole {
			case 1:
				fmt.Println("\nYou have been verified as an Administrator.")
			case 2:
				fmt.Println("\nYou have been verified as a Company user.")
			case 3:
				fmt.Println("\nYou have been verified as an Individual user.")
			default:
				fmt.Printf("\nYour account has been verified with role: %d\n", responseData.UserRole)
			}

			fmt.Println("\nYou can now log in to access your account.")
		},
		Args: cobra.MaximumNArgs(1), // Allow at most one argument (for the code)
	}

	// Define command flags
	cmd.Flags().StringVarP(&verificationCode, "code", "c", "", "Email verification code received after registration")

	return cmd
}
