// native/desktop/maplefile-cli/cmd/logout/logout.go
package logout

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
)

func LogoutCmd(logoutService auth.LogoutService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "logout",
		Short: "Log out the current user",
		Long: `Log out the current user by clearing stored authentication credentials.

This command will:
- Clear the stored access and refresh tokens
- Remove the saved user email from local configuration
- Require a new login to access protected features

Examples:
  # Log out the current user
  maplefile-cli logout
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging out...")

			// Create context
			ctx := context.Background()

			// Call the service to perform logout
			if err := logoutService.Logout(ctx); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ Logout successful!")
			fmt.Println("You have been logged out. Please use 'maplefile-cli requestloginott' to log in again.")
		},
	}

	return cmd
}
