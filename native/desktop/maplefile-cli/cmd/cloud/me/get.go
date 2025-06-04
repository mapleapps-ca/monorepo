// native/desktop/maplefile-cli/cmd/me/get.go
package me

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_me "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/me"
)

// getCmd creates a command for getting user profile
func getCmd(
	getService svc_me.GetMeService,
	logger *zap.Logger,
) *cobra.Command {
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Get your user profile",
		Long: `
Get your current user profile information from the cloud.

This command retrieves your user profile data including personal information,
contact details, and preferences.

Examples:
  # Get your profile
  maplefile-cli me get

  # Get your profile with detailed information
  maplefile-cli me get --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Execute get operation
			output, err := getService.Get(cmd.Context())
			if err != nil {
				fmt.Printf("🐞 Error getting user profile: %v\n", err)
				if strings.Contains(err.Error(), "unauthorized") {
					fmt.Printf("❌ Error: You need to login first. Run 'maplefile-cli requestloginott' to start the login process.\n")
				} else {
					logger.Error("Failed to get user profile", zap.Error(err))
				}
				return
			}

			// Display profile
			profile := output.Profile
			fmt.Printf("✅ User Profile:\n\n")
			fmt.Printf("Name: %s\n", profile.Name)
			fmt.Printf("Email: %s\n", profile.Email)
			fmt.Printf("Phone: %s\n", profile.Phone)
			fmt.Printf("Country: %s\n", profile.Country)

			if profile.Region != "" {
				fmt.Printf("Region: %s\n", profile.Region)
			}
			if profile.City != "" {
				fmt.Printf("City: %s\n", profile.City)
			}
			if profile.PostalCode != "" {
				fmt.Printf("Postal Code: %s\n", profile.PostalCode)
			}
			if profile.AddressLine1 != "" {
				fmt.Printf("Address: %s\n", profile.AddressLine1)
				if profile.AddressLine2 != "" {
					fmt.Printf("         %s\n", profile.AddressLine2)
				}
			}

			fmt.Printf("Timezone: %s\n", profile.Timezone)
			fmt.Printf("Email Verified: %t\n", profile.WasEmailVerified)

			if verbose {
				fmt.Printf("\nAdditional Information:\n")
				fmt.Printf("User ID: %s\n", profile.ID.Hex())
				fmt.Printf("Role: %d\n", profile.Role)
				fmt.Printf("Status: %d\n", profile.Status)
				fmt.Printf("Lexical Name: %s\n", profile.LexicalName)
				fmt.Printf("Agree to Promotions: %t\n", profile.AgreePromotions)
				fmt.Printf("Agree to Tracking: %t\n", profile.AgreeToTrackingAcrossThirdPartyAppsAndServices)
				fmt.Printf("Created: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
			}

			logger.Info("User profile retrieved successfully",
				zap.String("email", profile.Email),
				zap.String("name", profile.Name))
		},
	}

	// Define command flags
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed profile information")

	return cmd
}
