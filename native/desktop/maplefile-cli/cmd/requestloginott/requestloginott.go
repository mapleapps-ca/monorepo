// monorepo/native/desktop/maplefile-cli/cmd/requestloginott/requestloginott.go
package requestloginott

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

func RequestLoginOneTimeTokenUserCmd(loginOTTService auth.LoginOTTService, logger *zap.Logger) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "requestloginott",
		Short: "Request one-time login token for email",
		Long: `
Command will execute login command and user will get credentials to make API calls to their account.

After registration and email verification, use this command to log in to your account.

Examples:
  # Request login OTT with email
  maplefile-cli requestloginott --email user@example.com
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Requesting login one-time token...")

			if email == "" {
				log.Fatal("Email is required")
			}

			// Create context
			ctx := context.Background()

			// Call our service
			err := loginOTTService.RequestLoginOTT(ctx, email)
			if err != nil {
				log.Fatalf("Error: %v", err)
				return
			}

			// Success message
			fmt.Println("\n✅ One-time login token request successful!")
			fmt.Println("Please check your email for a one-time token.")
			fmt.Println("\nOnce you receive the token, run the following command to verify it:")
			fmt.Printf("maplefile-cli verifyloginott --email %s --ott YOUR_TOKEN\n", email)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
