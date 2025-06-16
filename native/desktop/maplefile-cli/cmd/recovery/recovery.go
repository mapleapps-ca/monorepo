// cmd/recovery/recovery.go
package recovery

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/recovery"
)

// RecoveryCmd creates the recovery command group
func RecoveryCmd(
	recoveryService recovery.RecoveryService,
	recoveryKeyService recovery.RecoveryKeyService,
	recoveryCleanupService recovery.RecoveryCleanupService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "recovery",
		Short: "Account recovery commands",
		Long: `Commands for recovering your account using a recovery key.

The recovery system follows the E2EE (End-to-End Encryption) architecture:
- Your master key is encrypted with a recovery key
- The recovery key allows you to reset your password without losing access to encrypted data
- All encryption happens locally - the server never sees your keys`,
	}

	// Add subcommands
	cmd.AddCommand(startRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(verifyRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(completeRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(statusRecoveryCmd(recoveryService, logger))
	cmd.AddCommand(showRecoveryKeyCmd(recoveryKeyService, logger))
	cmd.AddCommand(regenerateRecoveryKeyCmd(recoveryKeyService, logger))

	return cmd
}

// startRecoveryCmd creates the command to start recovery
func startRecoveryCmd(recoveryService recovery.RecoveryService, logger *zap.Logger) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start account recovery",
		Long: `Start the account recovery process using your recovery key.

This initiates the recovery process by:
1. Requesting a recovery session from the server
2. Receiving an encrypted challenge that only your recovery key can decrypt
3. Preparing for password reset while maintaining E2EE

Example:
  # Start recovery with email
  maplefile-cli recovery start --email user@example.com`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate email
			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			fmt.Println("🔐 Starting account recovery...")
			fmt.Printf("📧 Email: %s\n\n", email)

			// Start recovery
			result, err := recoveryService.InitiateRecovery(ctx, email)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				if strings.Contains(err.Error(), "rate limit") {
					fmt.Println("\n💡 Too many attempts. Please wait before trying again.")
				}
				return
			}

			fmt.Println("✅ Recovery initiated successfully!")
			fmt.Printf("🔑 Session ID: %s\n", result.SessionID)
			fmt.Printf("⏰ Expires at: %s\n", result.ExpiresAt.Format("15:04:05"))
			fmt.Println("\n🔐 The server has sent an encrypted challenge.")
			fmt.Println("📋 Next step: Verify your recovery key")
			fmt.Printf("\n👉 Run: maplefile-cli recovery verify --session %s\n", result.SessionID)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the account (required)")
	cmd.MarkFlagRequired("email")

	return cmd
}

// verifyRecoveryCmd creates the command to verify recovery key
func verifyRecoveryCmd(recoveryService recovery.RecoveryService, logger *zap.Logger) *cobra.Command {
	var sessionID string
	var recoveryKeyFile string

	var cmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify your recovery key",
		Long: `Verify your recovery key to decrypt the challenge and proceed with recovery.

This step:
1. Uses your recovery key to decrypt the challenge from the server
2. Proves you own the recovery key without sending it to the server
3. Receives a recovery token to complete the password reset

You can provide the recovery key:
- From a file using --recovery-key-file
- Interactively when prompted

Example:
  # Verify with session ID (will prompt for recovery key)
  maplefile-cli recovery verify --session <session-id>

  # Verify with recovery key from file
  maplefile-cli recovery verify --session <session-id> --recovery-key-file ~/recovery.key`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate session ID
			if sessionID == "" {
				fmt.Println("❌ Error: session ID is required")
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

			// Clean the recovery key (remove any formatting)
			recoveryKey = cleanRecoveryKey(recoveryKey)

			fmt.Println("\n🔐 Verifying recovery key...")

			// Verify recovery
			result, err := recoveryService.VerifyRecoveryKey(ctx, sessionID, recoveryKey)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				if strings.Contains(err.Error(), "incorrect") || strings.Contains(err.Error(), "invalid") {
					fmt.Println("\n💡 Please check your recovery key and try again.")
				}
				return
			}

			fmt.Println("\n✅ Recovery key verified successfully!")
			fmt.Println("🔓 Your identity has been confirmed.")
			fmt.Printf("🎟️ Recovery token expires at: %s\n", result.ExpiresAt.Format("15:04:05"))
			fmt.Println("\n📋 Next step: Set a new password")
			fmt.Println("👉 Run: maplefile-cli recovery complete")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Recovery session ID (required)")
	cmd.Flags().StringVarP(&recoveryKeyFile, "recovery-key-file", "f", "", "Path to file containing recovery key")
	cmd.MarkFlagRequired("session")

	return cmd
}

// completeRecoveryCmd creates the command to complete recovery
func completeRecoveryCmd(recoveryService recovery.RecoveryService, logger *zap.Logger) *cobra.Command {
	var recoveryToken string
	var showNewKey bool
	var recoveryKey string

	var cmd = &cobra.Command{
		Use:   "complete",
		Short: "Complete recovery with new password",
		Long: `Complete the account recovery process by setting a new password.

This final step:
1. Sets your new password
2. Re-encrypts your master key with the new password
3. Generates a new recovery key for future use
4. Maintains all your encrypted data intact

Your encrypted files remain accessible because the master key stays the same -
only the password protecting it changes.

If your recovery session was interrupted, you may need to provide your recovery key again.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Check recovery status first
			status, err := recoveryService.GetRecoveryStatus(ctx)
			if err != nil {
				fmt.Printf("❌ Error checking recovery status: %v\n", err)
				return
			}

			if !status.InProgress {
				fmt.Println("❌ Error: no active recovery session found")
				fmt.Println("\n💡 Recovery flow:")
				fmt.Println("   1. maplefile-cli recovery start --email <email>")
				fmt.Println("   2. maplefile-cli recovery verify --session <session-id>")
				fmt.Println("   3. maplefile-cli recovery complete")
				fmt.Println("\n👉 Start recovery first: maplefile-cli recovery start --email <email>")
				return
			}

			if status.Stage != "verified" {
				fmt.Printf("❌ Error: recovery key not yet verified (current stage: %s)\n", status.Stage)
				if status.Stage == "initiated" {
					fmt.Printf("👉 Verify first: maplefile-cli recovery verify --session %s\n", status.SessionID)
				}
				return
			}

			fmt.Printf("📧 Completing recovery for: %s\n", status.Email)

			// Prompt for new password
			password, err := promptForNewPassword()
			if err != nil {
				fmt.Printf("❌ Error reading password: %v\n", err)
				return
			}

			// Validate password
			if err := validatePassword(password); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n🔐 Setting new password...")

			// Complete recovery
			result, err := recoveryService.CompleteRecovery(ctx, recoveryToken, password)
			if err != nil {
				// Check if this is a missing recovery data error and we can prompt for recovery key
				if strings.Contains(err.Error(), "recovery data not found") && recoveryKey == "" {
					fmt.Println("⚠️  Recovery data not found in memory. This can happen if the CLI was restarted.")
					fmt.Println("🔑 Please provide your recovery key to complete the process:")

					// Prompt for recovery key
					promptedRecoveryKey, keyErr := promptForRecoveryKey()
					if keyErr != nil {
						fmt.Printf("❌ Error reading recovery key: %v\n", keyErr)
						return
					}

					// Clean the recovery key
					cleanKey := cleanRecoveryKey(promptedRecoveryKey)

					// Try to re-verify with the recovery key to restore recovery data
					fmt.Println("🔄 Re-verifying recovery key to restore session data...")
					_, verifyErr := recoveryService.VerifyRecoveryKey(ctx, status.SessionID, cleanKey)
					if verifyErr != nil {
						fmt.Printf("❌ Failed to verify recovery key: %v\n", verifyErr)
						fmt.Println("\n💡 Please ensure you're using the correct recovery key.")
						return
					}

					fmt.Println("✅ Recovery key verified! Attempting to complete recovery again...")

					// Try completion again
					result, err = recoveryService.CompleteRecovery(ctx, recoveryToken, password)
					if err != nil {
						fmt.Printf("❌ Error: %v\n", err)
						return
					}
				} else {
					fmt.Printf("❌ Error: %v\n", err)

					// Provide helpful error messages based on error type
					if strings.Contains(err.Error(), "no active recovery session") {
						fmt.Println("\n💡 The recovery session may have expired or been completed.")
						fmt.Println("👉 Start a new recovery: maplefile-cli recovery start --email <email>")
					} else if strings.Contains(err.Error(), "not verified") {
						fmt.Println("\n💡 You need to verify your recovery key first.")
						fmt.Println("👉 Verify: maplefile-cli recovery verify --session <session-id>")
					} else if strings.Contains(err.Error(), "expired") {
						fmt.Println("\n💡 The recovery session has expired.")
						fmt.Println("👉 Start a new recovery: maplefile-cli recovery start --email <email>")
					}
					return
				}
			}

			fmt.Println("\n✅ Password reset successfully!")
			fmt.Printf("📧 Account recovered: %s\n", result.Email)

			if showNewKey && strings.Contains(result.Message, "recovery key:") {
				// Extract and display the new recovery key
				parts := strings.Split(result.Message, "recovery key: ")
				if len(parts) > 1 {
					fmt.Println("\n🔑 Your NEW recovery key:")
					fmt.Printf("\n%s\n", parts[1])
					fmt.Println("\n⚠️  IMPORTANT: Save this new recovery key!")
					fmt.Println("⚠️  Your old recovery key no longer works.")
				}
			} else {
				fmt.Println("\n💡 A new recovery key has been generated.")
				fmt.Println("👉 View it with: maplefile-cli recovery show-key")
			}

			fmt.Println("\n🎉 You can now log in with your new password!")
			fmt.Println("👉 Run: maplefile-cli login --email " + result.Email)
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&recoveryToken, "token", "", "Recovery token (if you have it)")
	cmd.Flags().BoolVar(&showNewKey, "show-new-key", true, "Display the new recovery key after reset")
	cmd.Flags().StringVar(&recoveryKey, "recovery-key", "", "Recovery key (if recovery data was lost)")

	return cmd
}

// statusRecoveryCmd creates the command to check recovery status
func statusRecoveryCmd(recoveryService recovery.RecoveryService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Check recovery session status",
		Long:  `Check if there is an active recovery session and its current stage.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			status, err := recoveryService.GetRecoveryStatus(ctx)
			if err != nil {
				fmt.Printf("❌ Error checking recovery status: %v\n", err)
				return
			}

			if !status.InProgress {
				fmt.Println("ℹ️  No active recovery session")
				fmt.Println("\n👉 Start recovery with: maplefile-cli recovery start --email <email>")
				return
			}

			fmt.Println("✅ Active recovery session found:")
			fmt.Printf("📧 Email: %s\n", status.Email)
			fmt.Printf("🔑 Session ID: %s\n", status.SessionID)
			fmt.Printf("📊 Stage: %s\n", formatStage(status.Stage))
			if status.ExpiresAt != nil {
				fmt.Printf("⏰ Expires at: %s\n", status.ExpiresAt.Format("15:04:05"))
			}

			// Show next steps based on stage
			fmt.Println("\n📋 Next step:")
			switch status.Stage {
			case "initiated":
				fmt.Printf("👉 Verify your recovery key: maplefile-cli recovery verify --session %s\n", status.SessionID)
			case "verified":
				fmt.Println("👉 Complete recovery: maplefile-cli recovery complete")
			case "completed":
				fmt.Println("✅ Recovery completed! You can now log in with your new password.")
			}
		},
	}

	return cmd
}

// showRecoveryKeyCmd creates the command to show the recovery key
func showRecoveryKeyCmd(recoveryKeyService recovery.RecoveryKeyService, logger *zap.Logger) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "show-key",
		Short: "Display your recovery key",
		Long: `Display your recovery key for safekeeping.

This command:
1. Requires your current password to decrypt the recovery key
2. Shows your recovery key in a format easy to save
3. Helps you backup your key securely

⚠️  IMPORTANT: Store your recovery key in a safe place!`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			fmt.Println("🔐 Retrieving your recovery key...")
			fmt.Println("🔑 You'll need to enter your password to decrypt it.\n")

			// Prompt for password
			fmt.Print("Enter your password: ")
			password, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				fmt.Printf("❌ Error reading password: %v\n", err)
				return
			}

			// Get recovery key
			result, err := recoveryKeyService.ShowRecoveryKey(ctx, email, string(password))
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ Your recovery key:")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("\n%s\n\n", result.RecoveryKey)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			fmt.Printf("\n📅 Created: %s\n", result.CreatedAt)
			fmt.Println("\n⚠️  " + result.Instructions)
			fmt.Println("\n💡 Tips for storing your recovery key:")
			fmt.Println("   • Write it down and store in a safe place")
			fmt.Println("   • Save in a password manager")
			fmt.Println("   • Store in a bank safety deposit box")
			fmt.Println("   • Do NOT store it with your password")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.MarkFlagRequired("email")

	return cmd
}

// regenerateRecoveryKeyCmd creates the command to regenerate recovery key
func regenerateRecoveryKeyCmd(recoveryKeyService recovery.RecoveryKeyService, logger *zap.Logger) *cobra.Command {
	var email string
	var confirm bool

	var cmd = &cobra.Command{
		Use:   "regenerate-key",
		Short: "Generate a new recovery key",
		Long: `Generate a new recovery key, replacing the old one.

This command:
1. Requires your current password
2. Generates a new recovery key
3. Invalidates your old recovery key
4. Maintains all your encrypted data

Use this if you suspect your recovery key has been compromised.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			if !confirm {
				fmt.Println("⚠️  This will invalidate your current recovery key!")
				fmt.Print("Are you sure you want to continue? (y/N): ")

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.ToLower(strings.TrimSpace(response))

				if response != "y" && response != "yes" {
					fmt.Println("❌ Operation cancelled")
					return
				}
			}

			fmt.Println("\n🔐 Generating new recovery key...")
			fmt.Println("🔑 You'll need to enter your password.\n")

			// Prompt for password
			fmt.Print("Enter your password: ")
			password, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				fmt.Printf("❌ Error reading password: %v\n", err)
				return
			}

			// Generate new recovery key
			result, err := recoveryKeyService.GenerateNewRecoveryKey(ctx, email, string(password))
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			fmt.Println("\n✅ New recovery key generated successfully!")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("\n%s\n\n", result.RecoveryKey)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			fmt.Printf("\n📅 Created: %s\n", result.CreatedAt)
			fmt.Println("\n⚠️  " + result.Instructions)
			fmt.Println("⚠️  Your old recovery key has been invalidated!")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().BoolVar(&confirm, "yes", false, "Skip confirmation prompt")
	cmd.MarkFlagRequired("email")

	return cmd
}

// Helper functions

// promptForRecoveryKey prompts the user to enter their recovery key
func promptForRecoveryKey() (string, error) {
	fmt.Println("Enter your recovery key:")
	fmt.Println("(You can paste the entire key, including any formatting)")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	recoveryKey, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(recoveryKey), nil
}

// cleanRecoveryKey removes formatting from a recovery key
func cleanRecoveryKey(key string) string {
	// Remove common formatting characters
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "\n", "")
	key = strings.ReplaceAll(key, "\r", "")
	key = strings.ReplaceAll(key, "\t", "")

	return strings.TrimSpace(key)
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

// validatePassword validates a password meets minimum requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more validation rules as needed
	if strings.TrimSpace(password) != password {
		return fmt.Errorf("password cannot start or end with whitespace")
	}

	return nil
}

// formatStage returns a human-readable stage description
func formatStage(stage string) string {
	switch stage {
	case "initiated":
		return "Recovery started, waiting for key verification"
	case "verified":
		return "Recovery key verified, ready to set new password"
	case "completed":
		return "Recovery completed"
	default:
		return stage
	}
}

// UnifiedRecoveryCmd creates a single command that handles the entire recovery flow
func UnifiedRecoveryCmd(
	recoveryService recovery.RecoveryService,
	recoveryKeyService recovery.RecoveryKeyService,
	logger *zap.Logger,
) *cobra.Command {
	var email string
	var recoveryKeyFile string
	var skipVerify bool
	var skipComplete bool

	var cmd = &cobra.Command{
		Use:   "recover",
		Short: "Recover your account (unified flow)",
		Long: `Recover your account using your recovery key in a single command.

This unified command guides you through the entire recovery process:
1. Initiates recovery with your email
2. Verifies your recovery key
3. Sets a new password

The E2EE architecture ensures:
- Your master key remains unchanged
- All encrypted data stays accessible
- Only the password protecting your master key changes

Example:
  # Complete recovery flow (interactive)
  maplefile-cli recover --email user@example.com

  # With recovery key from file
  maplefile-cli recover --email user@example.com --recovery-key-file ~/recovery.key`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate email
			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			fmt.Printf("🔐 Starting account recovery for: %s\n\n", email)

			// STEP 1: Initiate recovery
			if !skipVerify {
				fmt.Println("📧 Step 1/3: Initiating recovery...")

				result, err := recoveryService.InitiateRecovery(ctx, email)
				if err != nil {
					fmt.Printf("❌ Failed to initiate recovery: %v\n", err)
					return
				}

				fmt.Println("✅ Recovery initiated!")
				fmt.Printf("🔑 Session ID: %s\n\n", result.SessionID)

				// STEP 2: Verify recovery key
				var recoveryKey string
				if recoveryKeyFile != "" {
					keyBytes, err := os.ReadFile(recoveryKeyFile)
					if err != nil {
						fmt.Printf("❌ Error reading recovery key file: %v\n", err)
						return
					}
					recoveryKey = strings.TrimSpace(string(keyBytes))
				} else {
					fmt.Println("🔑 Step 2/3: Enter your recovery key")
					recoveryKey, err = promptForRecoveryKey()
					if err != nil {
						fmt.Printf("❌ Error reading recovery key: %v\n", err)
						return
					}
				}

				recoveryKey = cleanRecoveryKey(recoveryKey)
				fmt.Println("\n🔐 Verifying recovery key...")

				_, err = recoveryService.VerifyRecoveryKey(ctx, result.SessionID, recoveryKey)
				if err != nil {
					fmt.Printf("❌ Failed to verify recovery key: %v\n", err)
					return
				}

				fmt.Println("✅ Recovery key verified!\n")
			}

			// STEP 3: Complete recovery
			if !skipComplete {
				fmt.Println("🔐 Step 3/3: Set new password")

				password, err := promptForNewPassword()
				if err != nil {
					fmt.Printf("❌ Error reading password: %v\n", err)
					return
				}

				if err := validatePassword(password); err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					return
				}

				fmt.Println("\n🔄 Completing recovery...")

				result, err := recoveryService.CompleteRecovery(ctx, "", password)
				if err != nil {
					fmt.Printf("❌ Failed to complete recovery: %v\n", err)
					return
				}

				fmt.Println("\n🎉 Account recovery successful!")
				fmt.Printf("✅ Password reset for: %s\n", result.Email)

				// Show new recovery key info
				if strings.Contains(result.Message, "recovery key:") {
					parts := strings.Split(result.Message, "recovery key: ")
					if len(parts) > 1 {
						fmt.Println("\n🔑 Your NEW recovery key:")
						fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
						fmt.Printf("\n%s\n\n", parts[1])
						fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
						fmt.Println("\n⚠️  Save this new recovery key - your old one no longer works!")
					}
				}

				fmt.Println("\n✅ You can now log in with your new password!")
				fmt.Printf("👉 Run: maplefile-cli login --email %s\n", email)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&recoveryKeyFile, "recovery-key-file", "f", "", "Path to recovery key file")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip to password reset (if already verified)")
	cmd.Flags().BoolVar(&skipComplete, "skip-complete", false, "Stop after verification")
	cmd.MarkFlagRequired("email")

	return cmd
}
