// monorepo/native/desktop/papercloud-cli/cmd/requestloginott/requestloginott.go
package requestloginott

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func RequestLoginOneTimeTokenUserCmd() *cobra.Command {
	var email, password string

	var cmd = &cobra.Command{
		Use:   "requestloginott",
		Short: "Request one-time login token for email",
		Long: `
Command will execute login command and user will get credentials to make API calls to their account.

After registration and email verification, use this command to log in to your account.

Examples:
  # Login with email and password
  go run main.go requestloginott --email user@example.com

`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging in...")

			if email == "" {
				log.Fatal("Email is required")
			}

			// Sanitize inputs
			email = strings.ToLower(strings.TrimSpace(email))

			// client := createE2EEClient()

			// // Call the Login method
			// err := client.RequestLoginOTT(email)
			// if err != nil {
			// 	log.Fatalf("Failed to login: %v", err)
			// }

			fmt.Println("Please check your email for a one-time token and enter it when prompted via the `verifyloginott` command.")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (will prompt if not provided)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
