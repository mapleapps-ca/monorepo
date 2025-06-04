// native/desktop/maplefile-cli/cmd/me/update.go
package me

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/me"
	svc_me "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/me"
)

// updateCmd creates a command for updating user profile
func updateCmd(
	updateService svc_me.UpdateMeService,
	logger *zap.Logger,
) *cobra.Command {
	var (
		email                                          string
		firstName                                      string
		lastName                                       string
		phone                                          string
		country                                        string
		region                                         string
		timezone                                       string
		agreePromotions                                bool
		agreeToTrackingAcrossThirdPartyAppsAndServices bool
	)

	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update your user profile",
		Long: `
Update your user profile information in the cloud.

This command allows you to update your personal information, contact details,
and preferences. All fields are required except region.

Examples:
  # Update your profile
  maplefile-cli me update --email john@example.com --first-name John --last-name Doe --phone "+1234567890" --country USA --timezone "America/New_York"

  # Update with promotional preferences
  maplefile-cli me update --email john@example.com --first-name John --last-name Doe --phone "+1234567890" --country USA --timezone "America/New_York" --agree-promotions --agree-tracking
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if email == "" {
				fmt.Println("🐞 Error: Email is required.")
				fmt.Println("Use --email flag to specify your email address.")
				return
			}

			if firstName == "" {
				fmt.Println("🐞 Error: First name is required.")
				fmt.Println("Use --first-name flag to specify your first name.")
				return
			}

			if lastName == "" {
				fmt.Println("🐞 Error: Last name is required.")
				fmt.Println("Use --last-name flag to specify your last name.")
				return
			}

			if phone == "" {
				fmt.Println("🐞 Error: Phone is required.")
				fmt.Println("Use --phone flag to specify your phone number.")
				return
			}

			if country == "" {
				fmt.Println("🐞 Error: Country is required.")
				fmt.Println("Use --country flag to specify your country.")
				return
			}

			if timezone == "" {
				fmt.Println("🐞 Error: Timezone is required.")
				fmt.Println("Use --timezone flag to specify your timezone (e.g., 'America/New_York', 'Europe/London').")
				return
			}

			// Create the service input
			input := &me.UpdateInput{
				Email:           email,
				FirstName:       firstName,
				LastName:        lastName,
				Phone:           phone,
				Country:         country,
				Region:          region,
				Timezone:        timezone,
				AgreePromotions: agreePromotions,
				AgreeToTrackingAcrossThirdPartyAppsAndServices: agreeToTrackingAcrossThirdPartyAppsAndServices,
			}

			// Execute update operation
			output, err := updateService.Update(cmd.Context(), input)
			if err != nil {
				fmt.Printf("🐞 Error updating user profile: %v\n", err)
				if strings.Contains(err.Error(), "unauthorized") {
					fmt.Printf("❌ Error: You need to login first. Run 'maplefile-cli requestloginott' to start the login process.\n")
				} else if strings.Contains(err.Error(), "already in use") {
					fmt.Printf("❌ Error: Email address is already in use by another user.\n")
				} else {
					logger.Error("Failed to update user profile",
						zap.String("email", email),
						zap.Error(err))
				}
				return
			}

			// Display success message
			profile := output.Profile
			fmt.Printf("✅ Successfully updated user profile!\n\n")
			fmt.Printf("Profile Details:\n")
			fmt.Printf("  Name: %s\n", profile.Name)
			fmt.Printf("  Email: %s\n", profile.Email)
			fmt.Printf("  Phone: %s\n", profile.Phone)
			fmt.Printf("  Country: %s\n", profile.Country)
			if profile.Region != "" {
				fmt.Printf("  Region: %s\n", profile.Region)
			}
			fmt.Printf("  Timezone: %s\n", profile.Timezone)
			fmt.Printf("  Agree to Promotions: %t\n", profile.AgreePromotions)
			fmt.Printf("  Agree to Tracking: %t\n", profile.AgreeToTrackingAcrossThirdPartyAppsAndServices)

			logger.Info("User profile updated successfully",
				zap.String("email", profile.Email),
				zap.String("name", profile.Name))
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&email, "email", "", "Email address (required)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name (required)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name (required)")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number (required)")
	cmd.Flags().StringVar(&country, "country", "", "Country (required)")
	cmd.Flags().StringVar(&region, "region", "", "Region/State (optional)")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone (required, e.g., 'America/New_York')")
	cmd.Flags().BoolVar(&agreePromotions, "agree-promotions", false, "Agree to receive promotional communications")
	cmd.Flags().BoolVar(&agreeToTrackingAcrossThirdPartyAppsAndServices, "agree-tracking", false, "Agree to tracking across third-party apps and services")

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("first-name")
	cmd.MarkFlagRequired("last-name")
	cmd.MarkFlagRequired("phone")
	cmd.MarkFlagRequired("country")
	cmd.MarkFlagRequired("timezone")

	return cmd
}
