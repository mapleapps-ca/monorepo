// monorepo/native/desktop/maplefile-cli/cmd/completelogin/completelogin.go
package completelogin

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

func CompleteLoginCmd(completeLoginService auth.CompleteLoginService, logger *zap.Logger) *cobra.Command {
	var email, password string
	var debugMode bool

	var cmd = &cobra.Command{
		Use:   "completelogin",
		Short: "Finish the login process",
		Long: `
After verifying the one-time token, use this command to complete the login process.

Examples:
  # Complete login with email and password
  maplefile-cli completelogin --email user@example.com --password yourpassword
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Completing login...")

			if email == "" {
				log.Fatal("Email is required")
			}

			if password == "" {
				log.Fatal("Password is required. Use --password flag (not secure for production use)")
			}

			// Create context
			ctx := context.Background()

			// Enable debug logging if requested
			if debugMode {
				// This would typically be handled by a proper logging config
				// but we'll simulate it here for simplicity
				logger.Info("Debug mode enabled")
			}

			// Call the service to complete login
			tokenResp, err := completeLoginService.CompleteLogin(ctx, email, password)
			if err != nil {
				log.Fatalf("Login failed: %v", err)
				return
			}

			// For debugging purposes
			fmt.Println("\n✅ Login successful!")
			fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
			fmt.Printf("Access Token Expires: %s\n", tokenResp.AccessTokenExpiryTime.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Printf("Refresh Token Expires: %s\n", tokenResp.RefreshTokenExpiryTime.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Println("\n✅ Saved email to our local configuration!")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (required)")
	cmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug output")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
