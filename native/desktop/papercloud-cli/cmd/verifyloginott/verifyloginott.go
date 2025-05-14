// monorepo/native/desktop/papercloud-cli/cmd/verifyloginott/verifyloginott.go
package verifyloginott

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func VerifyLoginOneTimeTokenUserCmd() *cobra.Command {
	var email, ott string

	var cmd = &cobra.Command{
		Use:   "verifyloginott",
		Short: "Verify one-time login token for email",
		Long: `
Verify the OTT and get the encrypted keys and challenge use this command to log in to your account.

Examples:
  # Login with email and password
  go run main.go verifyloginott --email user@example.com --ott=000000
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging in...")

			if email == "" {
				log.Fatal("Email is required")
			}

			// Sanitize inputs
			email = strings.ToLower(strings.TrimSpace(email))
			ott = strings.ToLower(strings.TrimSpace(ott))

			if email == "" {
				log.Fatal("Login: Email input cannot be empty")
			}
			if ott == "" {
				log.Fatal("Login: OTT input cannot be empty")
			}

			// client := createE2EEClient()

			// // Step 2: Verify the OTT and get the encrypted keys and challenge
			// ottResponse, err := client.VerifyLoginOTT(email, ott)
			// if err != nil {
			// 	log.Fatalf("Login: Failed for email %s: %v", email, err)
			// }

			// if err := preferences.SetVerifyOTTResponse(
			// 	ottResponse.EncryptedMasterKey,
			// 	ottResponse.Salt,
			// 	ottResponse.EncryptedChallenge,
			// 	ottResponse.PublicKey,
			// 	ottResponse.EncryptedPrivateKey,
			// 	ottResponse.ChallengeID,
			// ); err != nil {
			// 	log.Fatalf("Login: Failed to save %s: %v", email, err)
			// }

			fmt.Println("Please check your email for a one-time token and enter it when prompted via the `verifyloginott` command.")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&ott, "ott", "o", "", "One-time token sent by the backend")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
