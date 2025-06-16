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
		Short: "Force refresh current authenticated user's token",
		Long: `
Forces a refresh of the current authenticated user's access token by using
the refresh token. This is useful when the access token has expired but
the refresh token is still valid.

For accounts with encrypted tokens (E2EE), you may need to provide your password:

Examples:
  # Try to refresh tokens (works for plaintext tokens)
  maplefile-cli refreshtoken

  # Refresh tokens with password prompt (for encrypted tokens)
  maplefile-cli refreshtoken --prompt-password

  # Refresh tokens with inline password (not recommended for security)
  maplefile-cli refreshtoken --password YOUR_PASSWORD

Note: If your account uses end-to-end encryption, the refresh process may
require your password to decrypt the new tokens.
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔄 Refreshing authentication token...")

			ctx := context.Background()

			// Handle password input if needed
			finalPassword := password
			if promptPassword {
				fmt.Print("🔐 Enter your password (for encrypted token decryption): ")
				passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println() // New line after password input

				if err != nil {
					fmt.Printf("❌ Error reading password: %v\n", err)
					return
				}
				finalPassword = string(passwordBytes)
			}

			// Execute the refresh
			var err error
			if finalPassword != "" {
				// Use password-based refresh for encrypted tokens
				err = refreshTokenUseCase.ExecuteWithPassword(ctx, finalPassword)
			} else {
				// Try standard refresh first (for plaintext tokens)
				err = refreshTokenUseCase.Execute(ctx)
			}

			if err != nil {
				// Check if this is the specific encrypted token error
				if err.Error() == "encrypted tokens detected - password required. Use 'maplefile-cli refreshtoken --password' or login again" {
					fmt.Println("❌ Your account uses encrypted tokens which require a password for refresh.")
					fmt.Println("💡 Try one of these options:")
					fmt.Println("   maplefile-cli refreshtoken --prompt-password")
					fmt.Println("   maplefile-cli login --email your@email.com")
					return
				}

				log.Fatalf("Failed to refresh token: %v", err)
				return
			}

			// Get the updated user data to display expiry information
			creds, err := configService.GetLoggedInUserCredentials(ctx)
			if err != nil {
				fmt.Printf("⚠️  Token refreshed but failed to get updated credentials: %v\n", err)
				return
			}

			fmt.Println("\n✅ Authentication tokens refreshed successfully!")
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

	return cmd
}
