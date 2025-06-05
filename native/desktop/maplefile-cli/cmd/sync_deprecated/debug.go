// native/desktop/maplefile-cli/cmd/sync/debug.go
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
		Long: `Diagnose common sync issues and provide recommendations.

This command will check:
- Authentication status and password validation
- Network connectivity to cloud backend
- Sync state consistency
- Common configuration issues

Examples:
  # Run full diagnostic
  maplefile-cli sync debug --all

  # Check only authentication
  maplefile-cli sync debug --auth

  # Check with password validation
  maplefile-cli sync debug --auth --password mypassword`,
		Run: func(cmd *cobra.Command, args []string) {
			// If --all flag is used, enable all checks
			if cmd.Flag("all").Changed {
				checkAuth = true
				checkNetwork = true
				checkSyncState = true
			}

			// If no specific checks requested, default to all
			if !checkAuth && !checkNetwork && !checkSyncState {
				checkAuth = true
				checkNetwork = true
				checkSyncState = true
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

			// Clear password from memory
			password = ""

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

			if len(result.Issues) > 0 {
				fmt.Printf("\n⚠️  Issues Found (%d):\n", len(result.Issues))
				for i, issue := range result.Issues {
					fmt.Printf("   %d. %s\n", i+1, issue)
				}
			}

			if len(result.Recommendations) > 0 {
				fmt.Printf("\n💡 Recommendations (%d):\n", len(result.Recommendations))
				for i, rec := range result.Recommendations {
					fmt.Printf("   %d. %s\n", i+1, rec)
				}
			}

			if len(result.Issues) == 0 {
				fmt.Println("\n✅ No issues detected - sync should work properly!")
			}
		},
	}

	// Add command flags
	cmd.Flags().StringVarP(&password, "password", "p", "", "User password for testing decryption")
	cmd.Flags().BoolVar(&checkAuth, "auth", false, "Check authentication status")
	cmd.Flags().BoolVar(&checkNetwork, "network", false, "Check network connectivity")
	cmd.Flags().BoolVar(&checkSyncState, "sync-state", false, "Check sync state")
	cmd.Flags().Bool("all", false, "Run all diagnostic checks")

	return cmd
}
