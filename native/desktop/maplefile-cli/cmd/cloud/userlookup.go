// monorepo/native/desktop/maplefile-cli/cmd/cloud/userlookup.go
package cloud

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
)

func PublicUserLookupCmd(
	configService config.ConfigService,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
) *cobra.Command {
	var email string

	var cmd = &cobra.Command{
		Use:   "public-user-lookup",
		Short: "Lookup user",
		Long:  `Command will execute call to backend server to see if a particular email exists and if it does then return public information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing public user lookup...")

			// Validate required fields
			if email == "" {
				fmt.Println("❌ Error: email is required")
				return
			}

			// Create request DTO
			req := &publiclookupdto.PublicLookupRequestDTO{
				Email: email,
			}

			// Execute use case
			ctx := context.Background()
			response, err := getPublicLookupFromCloudUseCase.Execute(ctx, req)
			if err != nil {
				fmt.Printf("❌ Error performing lookup: %v\n", err)
				return
			}

			// Display the results
			fmt.Println("✅ User found!")
			fmt.Printf("User ID: %s\n", response.UserID)
			fmt.Printf("Email: %s\n", response.Email)
			fmt.Printf("Name: %s\n", response.Name)
			fmt.Printf("PublicKey (Base64 encoded): %s\n", response.PublicKeyInBase64)
			fmt.Printf("VerificationID: %s\n", response.VerificationID)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}
