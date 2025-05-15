// monorepo/native/desktop/maplefile-cli/cmd/verifyloginott/verifyloginott.go
package verifyloginott

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

func VerifyLoginOneTimeTokenUserCmd(loginOTTVerificationService auth.LoginOTTVerificationService, logger *zap.Logger) *cobra.Command {
	var email, ott string

	var cmd = &cobra.Command{
		Use:   "verifyloginott",
		Short: "Verify one-time login token for email",
		Long: `
Verify the OTT and get the encrypted keys and challenge for login.

After requesting a one-time token with requestloginott, use this command to verify it.

Examples:
  # Verify with email and one-time token
  maplefile-cli verifyloginott --email user@example.com --ott 123456
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Verifying one-time login token...")

			// Validate required fields
			if email == "" {
				log.Fatal("Email is required")
			}
			if ott == "" {
				log.Fatal("One-time token (OTT) is required")
			}

			// Create context
			ctx := context.Background()

			// Call our service
			err := loginOTTVerificationService.VerifyLoginOTT(ctx, email, ott)
			if err != nil {
				log.Fatalf("Error: %v", err)
				return
			}

			// Success message
			fmt.Println("\n✅ One-time token verified successfully!")
			fmt.Println("You can now proceed to complete the login process with:")
			fmt.Printf("maplefile-cli completelogin --email %s\n", email)
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
