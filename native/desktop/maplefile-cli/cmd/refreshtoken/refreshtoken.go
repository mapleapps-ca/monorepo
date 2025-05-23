// monorepo/native/desktop/maplefile-cli/cmd/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/tokenservice"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/refreshtoken"
)

// RefreshTokenCmd creates a new command for refreshing authentication tokens
func RefreshTokenCmd(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefreshSvc tokenservice.TokenRefreshService,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "refreshtoken",
		Short: "Force refresh current authenticated user's token",
		Long: `
Forces a refresh of the current authenticated user's access token by using
the refresh token. This is useful when the access token has expired but
the refresh token is still valid.

This command does not take any arguments. It uses the refresh token
stored locally for the currently authenticated user.

Example:
  maplefile-cli refreshtoken
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Refreshing authentication token...")

			ctx := context.Background()

			// Create the use case with injected dependencies
			useCase := refreshtoken.NewRefreshTokenUseCase(
				logger,
				configService,
				userRepo,
				tokenRefreshSvc,
			)

			// Execute the use case
			if err := useCase.Execute(ctx); err != nil {
				log.Fatalf("Failed to refresh token: %v", err)
				return
			}

			// Get the updated user data to display expiry information
			email, _ := configService.GetLoggedInUserEmail(ctx)
			userData, _ := userRepo.GetByEmail(ctx, email)

			fmt.Println("\n✅ Authentication tokens refreshed successfully!")
			fmt.Printf("Access Token expires: %s\n", userData.AccessTokenExpiryTime.Format(time.RFC3339))
			fmt.Printf("Refresh Token expires: %s\n", userData.RefreshTokenExpiryTime.Format(time.RFC3339))
		},
	}

	return cmd
}
