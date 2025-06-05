// cmd/login/login.go - NEW FILE
package login

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"

	svc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/authdto"
)

// LoginCmd creates a unified login command that handles the complete authentication flow
func LoginCmd(
	loginOTTService svc_authdto.LoginOTTService,
	loginOTTVerificationService svc_authdto.LoginOTTVerificationService,
	completeLoginService svc_authdto.CompleteLoginService,
	logger *zap.Logger,
) *cobra.Command {
	var email string
	var ott string
	var password string
	var skipOTTRequest bool
	var skipOTTVerification bool

	var cmd = &cobra.Command{
		Use:   "login",
		Short: "Log in to your MapleFile account",
		Long: `
Log in to your MapleFile account using the complete authentication flow.

This command will guide you through the entire login process:
1. Request a one-time token (OTT) to be sent to your email
2. Verify the OTT you received
3. Complete login with your password

You can also skip individual steps if you've already completed them manually.

Examples:
  # Complete login flow (recommended)
  maplefile-cli login --email user@example.com

  # Skip OTT request if you already have a token
  maplefile-cli login --email user@example.com --skip-ott-request

  # Skip to password entry if you've already verified your OTT
  maplefile-cli login --email user@example.com --skip-ott-verification

  # Provide all details upfront (non-interactive)
  maplefile-cli login --email user@example.com --ott 123456 --password mypassword
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate email is provided
			if email == "" {
				fmt.Println("❌ Error: Email is required")
				fmt.Println("Use: maplefile-cli login --email your@email.com")
				return
			}

			fmt.Printf("🔐 Starting login for: %s\n\n", email)

			// STEP 1: Request OTT (unless skipped)
			if !skipOTTRequest {
				fmt.Println("📧 Step 1/3: Requesting one-time token...")

				err := loginOTTService.RequestLoginOTT(ctx, email)
				if err != nil {
					fmt.Printf("❌ Failed to request login token: %v\n", err)
					fmt.Println("\n💡 Tip: You can run this step manually with:")
					fmt.Printf("   maplefile-cli request-login-token --email %s\n", email)
					return
				}

				fmt.Println("✅ One-time token sent to your email!")
				fmt.Println("📬 Please check your inbox for the verification code.\n")
			} else {
				fmt.Println("⏭️  Skipping OTT request (--skip-ott-request specified)\n")
			}

			// STEP 2: Get and verify OTT (unless skipped)
			if !skipOTTVerification {
				if ott == "" {
					fmt.Print("🔢 Step 2/3: Enter the verification code from your email: ")
					reader := bufio.NewReader(os.Stdin)
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("❌ Error reading input: %v\n", err)
						return
					}
					ott = strings.TrimSpace(input)
				}

				if ott == "" {
					fmt.Println("❌ Error: Verification code is required")
					fmt.Println("\n💡 Tip: You can run this step manually with:")
					fmt.Printf("   maplefile-cli verify-login-token --email %s --ott YOUR_CODE\n", email)
					return
				}

				fmt.Println("🔍 Verifying one-time token...")

				err := loginOTTVerificationService.VerifyLoginOTT(ctx, email, ott)
				if err != nil {
					fmt.Printf("❌ Failed to verify token: %v\n", err)
					if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
						fmt.Println("\n💡 Tips:")
						fmt.Println("   • Check that you entered the code correctly")
						fmt.Println("   • Codes expire after a few minutes - request a new one if needed")
						fmt.Printf("   • Request new token: maplefile-cli request-login-token --email %s\n", email)
					}
					return
				}

				fmt.Println("✅ Token verified successfully!\n")
			} else {
				fmt.Println("⏭️  Skipping OTT verification (--skip-ott-verification specified)\n")
			}

			// STEP 3: Get password and complete login
			if password == "" {
				fmt.Print("🔐 Step 3/3: Enter your password: ")
				passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println() // New line after password input

				if err != nil {
					fmt.Printf("❌ Error reading password: %v\n", err)
					return
				}
				password = string(passwordBytes)
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required")
				fmt.Println("\n💡 Tip: You can run this step manually with:")
				fmt.Printf("   maplefile-cli complete-login --email %s --password YOUR_PASSWORD\n", email)
				return
			}

			fmt.Println("🔄 Completing login...")

			tokenResp, err := completeLoginService.CompleteLogin(ctx, email, password)
			if err != nil {
				fmt.Printf("❌ Login failed: %v\n", err)
				if strings.Contains(err.Error(), "password") {
					fmt.Println("\n💡 Tips:")
					fmt.Println("   • Double-check your password")
					fmt.Println("   • Use account recovery if you've forgotten your password")
					fmt.Println("   • Run: maplefile-cli recovery start --email " + email)
				}
				return
			}

			// Success!
			fmt.Println("\n🎉 Login successful!")
			fmt.Printf("✅ Access token expires: %s\n", tokenResp.AccessTokenExpiryTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Refresh token expires: %s\n", tokenResp.RefreshTokenExpiryTime.Format("2006-01-02 15:04:05"))

			// Show next steps
			fmt.Println("\n🚀 You're now logged in! Try these commands:")
			fmt.Println("   maplefile-cli me                    # View your profile")
			fmt.Println("   maplefile-cli collections list      # List your collections")
			fmt.Println("   maplefile-cli sync                  # Sync with cloud")

			logger.Info("User logged in successfully",
				zap.String("email", email),
				zap.Time("login_time", time.Now()))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVar(&ott, "ott", "", "One-time token (if you already have it)")
	cmd.Flags().StringVar(&password, "password", "", "Password (not recommended for security)")
	cmd.Flags().BoolVar(&skipOTTRequest, "skip-ott-request", false, "Skip requesting OTT (if already requested)")
	cmd.Flags().BoolVar(&skipOTTVerification, "skip-ott-verification", false, "Skip OTT verification (if already verified)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}

// Individual command functions (kept for manual use)

// RequestLoginTokenCmd - Individual command for step 1
func RequestLoginTokenCmd(loginOTTService svc_authdto.LoginOTTService, logger *zap.Logger) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "request-login-token",
		Short: "Request one-time login token (manual step 1/3)",
		Long: `
Request a one-time login token to be sent to your email.

This is step 1 of the manual login process. For most users, the unified 'login'
command is recommended instead.

Manual login process:
1. maplefile-cli request-login-token --email user@example.com
2. maplefile-cli verify-login-token --email user@example.com --ott CODE
3. maplefile-cli complete-login --email user@example.com --password PASSWORD

Examples:
  # Request login token
  maplefile-cli request-login-token --email user@example.com
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if email == "" {
				fmt.Println("❌ Error: Email is required")
				return
			}

			fmt.Println("📧 Requesting one-time login token...")

			err := loginOTTService.RequestLoginOTT(ctx, email)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ One-time login token request successful!")
			fmt.Println("📬 Please check your email for a one-time token.")
			fmt.Println("\n🔗 Next step:")
			fmt.Printf("   maplefile-cli verify-login-token --email %s --ott YOUR_TOKEN\n", email)
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.MarkFlagRequired("email")

	return cmd
}

// VerifyLoginTokenCmd - Individual command for step 2
func VerifyLoginTokenCmd(loginOTTVerificationService svc_authdto.LoginOTTVerificationService, logger *zap.Logger) *cobra.Command {
	var email, ott string

	var cmd = &cobra.Command{
		Use:   "verify-login-token",
		Short: "Verify one-time login token (manual step 2/3)",
		Long: `
Verify the one-time login token you received via email.

This is step 2 of the manual login process. For most users, the unified 'login'
command is recommended instead.

Examples:
  # Verify login token
  maplefile-cli verify-login-token --email user@example.com --ott 123456
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if email == "" || ott == "" {
				fmt.Println("❌ Error: Both email and OTT are required")
				return
			}

			fmt.Println("🔍 Verifying one-time login token...")

			err := loginOTTVerificationService.VerifyLoginOTT(ctx, email, ott)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ One-time token verified successfully!")
			fmt.Println("🔗 Final step:")
			fmt.Printf("   maplefile-cli complete-login --email %s --password YOUR_PASSWORD\n", email)
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&ott, "ott", "o", "", "One-time token (required)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("ott")

	return cmd
}

// CompleteLoginCmd - Individual command for step 3
func CompleteLoginCmd(completeLoginService svc_authdto.CompleteLoginService, logger *zap.Logger) *cobra.Command {
	var email, password string
	var debugMode bool

	var cmd = &cobra.Command{
		Use:   "complete-login",
		Short: "Complete the login process (manual step 3/3)",
		Long: `
Complete the login process by providing your password.

This is step 3 of the manual login process. For most users, the unified 'login'
command is recommended instead.

Examples:
  # Complete login (will prompt for password)
  maplefile-cli complete-login --email user@example.com

  # Complete login with password (not recommended)
  maplefile-cli complete-login --email user@example.com --password mypassword
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if email == "" {
				fmt.Println("❌ Error: Email is required")
				return
			}

			// Prompt for password if not provided
			if password == "" {
				fmt.Print("🔐 Enter your password: ")
				passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println()

				if err != nil {
					fmt.Printf("❌ Error reading password: %v\n", err)
					return
				}
				password = string(passwordBytes)
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required")
				return
			}

			fmt.Println("🔄 Completing login...")

			if debugMode {
				logger.Info("Debug mode enabled")
			}

			tokenResp, err := completeLoginService.CompleteLogin(ctx, email, password)
			if err != nil {
				fmt.Printf("❌ Login failed: %v\n", err)
				return
			}

			fmt.Println("\n🎉 Login successful!")
			fmt.Printf("✅ Access token expires: %s\n", tokenResp.AccessTokenExpiryTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Refresh token expires: %s\n", tokenResp.RefreshTokenExpiryTime.Format("2006-01-02 15:04:05"))
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (will prompt if not provided)")
	cmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug output")
	cmd.MarkFlagRequired("email")

	return cmd
}
