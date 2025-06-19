// native/desktop/maplefile-cli/cmd/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/refreshtoken"
)

// RefreshTokenCmd creates a new command for refreshing authentication tokens
func RefreshTokenCmd(
	logger *zap.Logger,
	configService config.ConfigService,
	refreshTokenUseCase refreshtoken.RefreshTokenUseCase,
) *cobra.Command {
	var password string
	var promptPassword bool

	var cmd = &cobra.Command{
		Use:   "refreshtoken",
		Short: "Refresh current authenticated user's encrypted tokens",
		Long: `
Refreshes the current authenticated user's access token by using the refresh token.
Since all tokens are now encrypted, your password is required for decryption.

This is useful when the access token has expired but the refresh token is still valid.

Examples:
  # Refresh tokens with password prompt (recommended)
  maplefile-cli refreshtoken --prompt-password

  # Refresh tokens with inline password (not recommended for security)
  maplefile-cli refreshtoken --password YOUR_PASSWORD

Note: All tokens are now encrypted end-to-end, so your password is always required
to decrypt the new tokens received from the server.
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔄 Refreshing encrypted authentication tokens...")

			ctx := context.Background()

			// Handle password input
			finalPassword := password
			if promptPassword {
				fmt.Print("🔐 Enter your password (required for encrypted token decryption): ")
				passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println() // New line after password input

				if err != nil {
					fmt.Printf("❌ Error reading password: %v\n", err)
					return
				}
				finalPassword = string(passwordBytes)
			}

			// Validate that we have a password
			if finalPassword == "" {
				fmt.Println("❌ Password is required for encrypted token refresh.")
				fmt.Println("💡 Use one of these options:")
				fmt.Println("   maplefile-cli refreshtoken --prompt-password")
				fmt.Println("   maplefile-cli refreshtoken --password YOUR_PASSWORD")
				return
			}

			// Execute the refresh with password
			err := refreshTokenUseCase.ExecuteWithPassword(ctx, finalPassword)
			if err != nil {
				log.Fatalf("Failed to refresh encrypted tokens: %v", err)
				return
			}

			// Get the updated user data to display expiry information
			creds, err := configService.GetLoggedInUserCredentials(ctx)
			if err != nil {
				fmt.Printf("⚠️  Tokens refreshed but failed to get updated credentials: %v\n", err)
				return
			}

			fmt.Println("\n✅ Encrypted authentication tokens refreshed successfully!")
			fmt.Printf("📧 Email: %s\n", creds.Email)
			fmt.Printf("🔑 Access Token expires: %s\n", creds.AccessTokenExpiryTime.Format(time.RFC3339))
			fmt.Printf("🔄 Refresh Token expires: %s\n", creds.RefreshTokenExpiryTime.Format(time.RFC3339))

			// Show time until expiry
			timeUntilExpiry := time.Until(*creds.AccessTokenExpiryTime)
			if timeUntilExpiry > 0 {
				fmt.Printf("⏰ Access token valid for: %s\n", timeUntilExpiry.Round(time.Minute))
			} else {
				fmt.Println("⚠️  Access token has already expired!")
			}
		},
	}

	// Add password flags
	cmd.Flags().StringVar(&password, "password", "", "Password for encrypted token decryption (not recommended for security)")
	cmd.Flags().BoolVar(&promptPassword, "prompt-password", false, "Prompt for password (recommended for encrypted tokens)")

	// Make one of the password options required
	cmd.MarkFlagsOneRequired("password", "prompt-password")

	return cmd
}
