// monorepo/native/desktop/maplefile-cli/cmd/recovery/recovery.go
package recovery

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

// RecoveryCmd creates the recovery command group
func RecoveryCmd(recoveryService auth.RecoveryService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "recovery",
		Short: "Account recovery commands",
		Long:  `Commands for recovering your account using a recovery key`,
	}

	// Add subcommands
	cmd.AddCommand(startRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(completeRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(statusRecoveryCmd(recoveryService, logger))

	return cmd
}

// startRecoveryCmd creates the command to start recovery
func startRecoveryCmd(recoveryService auth.RecoveryService, logger *zap.Logger) *cobra.Command {
	var email string
	var recoveryKeyFile string

	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start account recovery",
		Long: `Start the account recovery process using your recovery key.

You can provide the recovery key in multiple ways:
1. From a file using --recovery-key-file
2. Interactively when prompted
3. Via stdin (for scripts)

Example:
  # Start recovery with email and key file
  maplefile-cli recovery start --email user@example.com --recovery-key-file ~/recovery.key

  # Start recovery and enter key interactively
  maplefile-cli recovery start --email user@example.com`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate email
			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			// Get recovery key
			var recoveryKey string
			var err error

			if recoveryKeyFile != "" {
				// Read from file
				keyBytes, err := os.ReadFile(recoveryKeyFile)
				if err != nil {
					fmt.Printf("❌ Error reading recovery key file: %v\n", err)
					return
				}
				recoveryKey = strings.TrimSpace(string(keyBytes))
			} else {
				// Prompt for recovery key
				recoveryKey, err = promptForRecoveryKey()
				if err != nil {
					fmt.Printf("❌ Error reading recovery key: %v\n", err)
					return
				}
			}

			// Validate recovery key format
			if recoveryKey == "" {
				fmt.Println("❌ Error: recovery key is required")
				return
			}

			// Check if it's base64 encoded
			_, err = base64.StdEncoding.DecodeString(recoveryKey)
			if err != nil {
				// Try URL-safe base64
				_, err = base64.RawURLEncoding.DecodeString(recoveryKey)
				if err != nil {
					fmt.Println("❌ Error: recovery key must be base64 encoded")
					return
				}
			}

			fmt.Println("🔐 Starting account recovery...")

			// Start recovery
			if err := recoveryService.InitiateRecovery(ctx, email, recoveryKey); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ Recovery initiated successfully!")
			fmt.Println("👉 Now run 'maplefile-cli recovery complete' to set a new password")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the account (required)")
	cmd.Flags().StringVarP(&recoveryKeyFile, "recovery-key-file", "f", "", "Path to file containing recovery key")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}

// completeRecoveryCmd creates the command to complete recovery
func completeRecoveryCmd(recoveryService auth.RecoveryService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "complete",
		Short: "Complete account recovery with new password",
		Long: `Complete the account recovery process by setting a new password.

This command must be run after 'recovery start' has been successful.
You will be prompted to enter and confirm your new password.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Check if recovery is in progress
			status, err := recoveryService.GetRecoveryStatus(ctx)
			if err != nil {
				fmt.Printf("❌ Error checking recovery status: %v\n", err)
				return
			}

			if !status.InProgress {
				fmt.Println("❌ Error: no recovery session in progress")
				fmt.Println("👉 Run 'maplefile-cli recovery start' first")
				return
			}

			fmt.Printf("📧 Completing recovery for: %s\n", status.Email)
			fmt.Printf("⏰ Session expires at: %s\n\n", status.ExpiresAt.Format("2006-01-02 15:04:05"))

			// Prompt for new password
			password, err := promptForNewPassword()
			if err != nil {
				fmt.Printf("❌ Error reading password: %v\n", err)
				return
			}

			// Validate password
			if len(password) < 8 {
				fmt.Println("❌ Error: password must be at least 8 characters long")
				return
			}

			fmt.Println("\n🔐 Setting new password...")

			// Complete recovery
			if err := recoveryService.SetNewPassword(ctx, password); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ Password reset successfully!")
			fmt.Println("🎉 Your account has been recovered")
			fmt.Println("👉 You can now log in with your new password")
		},
	}

	return cmd
}

// statusRecoveryCmd creates the command to check recovery status
func statusRecoveryCmd(recoveryService auth.RecoveryService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Check recovery session status",
		Long:  `Check if there is an active recovery session and its details.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			status, err := recoveryService.GetRecoveryStatus(ctx)
			if err != nil {
				fmt.Printf("❌ Error checking recovery status: %v\n", err)
				return
			}

			if !status.InProgress {
				fmt.Println("ℹ️  No active recovery session")
				return
			}

			fmt.Println("✅ Active recovery session found:")
			fmt.Printf("📧 Email: %s\n", status.Email)
			fmt.Printf("⏰ Expires at: %s\n", status.ExpiresAt.Format("2006-01-02 15:04:05"))
		},
	}

	return cmd
}

// promptForRecoveryKey prompts the user to enter their recovery key
func promptForRecoveryKey() (string, error) {
	fmt.Print("Enter your recovery key: ")

	reader := bufio.NewReader(os.Stdin)
	recoveryKey, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(recoveryKey), nil
}

// promptForNewPassword prompts the user to enter and confirm a new password
func promptForNewPassword() (string, error) {
	// First password
	fmt.Print("Enter new password: ")
	password1, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	// Confirm password
	fmt.Print("Confirm new password: ")
	password2, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	// Check if passwords match
	if string(password1) != string(password2) {
		return "", fmt.Errorf("passwords do not match")
	}

	return string(password1), nil
}

// ShowRecoveryKeyCmd creates the command to show the recovery key
func ShowRecoveryKeyCmd(recoveryKeyService auth.RecoveryKeyService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show-recovery-key",
		Short: "Display your recovery key",
		Long: `Display your recovery key for safekeeping.

This command requires you to authenticate with your password.
Store your recovery key in a safe place - you'll need it if you forget your password.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			fmt.Println("🔐 Retrieving your recovery key...")
			fmt.Println("⚠️  You will need to enter your password to decrypt the recovery key\n")

			// Prompt for password
			fmt.Print("Enter your password: ")
			password, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				fmt.Printf("❌ Error reading password: %v\n", err)
				return
			}

			// Get recovery key
			recoveryKey, err := recoveryKeyService.ShowRecoveryKey(ctx, string(password))
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ Your recovery key:")
			fmt.Printf("\n%s\n\n", recoveryKey)
			fmt.Println("⚠️  IMPORTANT: Store this key in a safe place!")
			fmt.Println("⚠️  You will need it to recover your account if you forget your password")
			fmt.Println("⚠️  Anyone with this key can reset your password")
		},
	}

	return cmd
}
