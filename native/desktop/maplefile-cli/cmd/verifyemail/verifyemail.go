// monorepo/native/desktop/maplefile-cli/cmd/verifyemail/verifyemail.go
package verifyemail

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	authService "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

func VerifyEmailCmd(emailVerificationService authService.EmailVerificationService, logger *zap.Logger) *cobra.Command {
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

			// Call our service
			result, err := emailVerificationService.VerifyEmail(ctx, verificationCode)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ Email verification successful!")

			if result.Message != "" {
				fmt.Printf("\n%s\n", result.Message)
			}

			// Display user role information
			switch result.UserRole {
			case 1:
				fmt.Println("\nYou have been verified as an Administrator.")
			case 2:
				fmt.Println("\nYou have been verified as a Company user.")
			case 3:
				fmt.Println("\nYou have been verified as an Individual user.")
			default:
				fmt.Printf("\nYour account has been verified with role: %d\n", result.UserRole)
			}

			fmt.Println("\nYou can now log in to access your account.")
		},
		Args: cobra.MaximumNArgs(1), // Allow at most one argument (for the code)
	}

	// Define command flags
	cmd.Flags().StringVarP(&verificationCode, "code", "c", "", "Email verification code received after registration")

	return cmd
}
