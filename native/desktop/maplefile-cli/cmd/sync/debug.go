// cmd/sync/debug.go - Clean debug command (unchanged functionality)
package sync

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// debugCmd creates a command for debugging sync operations
func debugCmd(
	syncDebugService svc_sync.SyncDebugService,
	logger *zap.Logger,
) *cobra.Command {
	var password string
	var checkAuth bool
	var checkNetwork bool
	var checkSyncState bool

	var cmd = &cobra.Command{
		Use:   "debug",
		Short: "Debug sync operations",
		Long: `
Diagnose common sync issues and provide recommendations.

This command checks:
  • Authentication status and password validation
  • Network connectivity to cloud backend
  • Sync state consistency
  • Common configuration issues

Examples:
  # Run full diagnostic
  maplefile-cli sync debug --password mypass

  # Check specific components
  maplefile-cli sync debug --auth --password mypass
  maplefile-cli sync debug --network
  maplefile-cli sync debug --sync-state --password mypass

  # Check all explicitly
  maplefile-cli sync debug --auth --network --sync-state --password mypass
`,
		Run: func(cmd *cobra.Command, args []string) {
			// If no specific checks requested, default to all
			if !checkAuth && !checkNetwork && !checkSyncState {
				checkAuth = true
				checkNetwork = true
				checkSyncState = true
			}

			// Validate password for auth/sync-state checks
			if (checkAuth || checkSyncState) && password == "" {
				fmt.Println("❌ Error: Password is required for authentication and sync state checks.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			fmt.Println("🔍 Running sync diagnostics...")

			// Create input for debug service
			input := &svc_sync.DebugSyncInput{
				CheckAuth:      checkAuth,
				CheckNetwork:   checkNetwork,
				CheckSyncState: checkSyncState,
				Password:       password,
			}

			// Execute diagnostics
			result, err := syncDebugService.DiagnoseSync(cmd.Context(), input)
			if err != nil {
				fmt.Printf("❌ Diagnostics failed: %v\n", err)
				return
			}

			// Display results
			fmt.Println("\n📊 Diagnostic Results:")

			if checkAuth {
				fmt.Printf("🔐 Authentication: %s\n", result.AuthStatus)
			}

			if checkNetwork {
				fmt.Printf("🌐 Network: %s\n", result.NetworkStatus)
			}

			if checkSyncState {
				fmt.Printf("📊 Sync State: %s\n", result.SyncStateStatus)
			}

			// Show issues if any
			if len(result.Issues) > 0 {
				fmt.Printf("\n⚠️  Issues Found (%d):\n", len(result.Issues))
				for i, issue := range result.Issues {
					fmt.Printf("   %d. %s\n", i+1, issue)
				}
			}

			// Show recommendations if any
			if len(result.Recommendations) > 0 {
				fmt.Printf("\n💡 Recommendations (%d):\n", len(result.Recommendations))
				for i, rec := range result.Recommendations {
					fmt.Printf("   %d. %s\n", i+1, rec)
				}
			}

			// Summary
			if len(result.Issues) == 0 {
				fmt.Println("\n✅ No issues detected - sync should work properly!")
				fmt.Println("💡 Try running: maplefile-cli sync --password PASSWORD")
			} else {
				fmt.Printf("\n🔧 Found %d issue(s) that may affect sync performance.\n", len(result.Issues))
				fmt.Println("💡 Address the recommendations above and try syncing again.")
			}

			logger.Info("Sync diagnostics completed",
				zap.Bool("authChecked", checkAuth),
				zap.Bool("networkChecked", checkNetwork),
				zap.Bool("syncStateChecked", checkSyncState),
				zap.Int("issuesFound", len(result.Issues)))
		},
	}

	// Define flags
	cmd.Flags().StringVar(&password, "password", "", "User password for testing decryption")
	cmd.Flags().BoolVar(&checkAuth, "auth", false, "Check authentication status")
	cmd.Flags().BoolVar(&checkNetwork, "network", false, "Check network connectivity")
	cmd.Flags().BoolVar(&checkSyncState, "sync-state", false, "Check sync state consistency")

	return cmd
}
