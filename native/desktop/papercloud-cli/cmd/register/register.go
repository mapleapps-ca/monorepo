// cmd/register/register.go
package register

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/register"
)

// RegisterCmd creates the register command for the CLI
func RegisterCmd(registerService register.RegisterService) *cobra.Command {
	var email, password, firstName, lastName, timezone, country, phone, betaAccessCode string
	var agreeTerms, agreePromotions, agreeTracking, skipRemoteRegistration bool
	var module int

	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Register an account on PaperCloud",
		Long: `Register a new user account in the system.

This command requires you to provide an email, password, first name, and last name.
You can optionally provide timezone, country, phone number, a beta access code,
specify agreement to terms, promotions, and tracking, and specify the registration module.

Registration information will be saved locally before being sent to the remote server.
Use the --skip-remote flag to only save locally without registering with the remote server.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if email == "" || password == "" || firstName == "" || lastName == "" {
				fmt.Println("Error: email, password, first name, and last name are required")
				return
			}

			if !agreeTerms {
				fmt.Println("Error: you must agree to the Terms of Service")
				return
			}

			// Set default module value if not specified
			if module <= 0 {
				module = 1 // Default to module 1 (PaperCloud)
			}

			// Generate E2EE fields
			fmt.Println("Generating secure cryptographic keys...")

			// Prepare the registration input
			input := register.RegisterUserInput{
				Email:           strings.ToLower(email),
				Password:        password,
				FirstName:       firstName,
				LastName:        lastName,
				Timezone:        timezone,
				Country:         country,
				Phone:           phone,
				BetaAccessCode:  betaAccessCode,
				AgreeTerms:      agreeTerms,
				AgreePromotions: agreePromotions,
				AgreeTracking:   agreeTracking,
				Module:          module,
				SkipRemoteReg:   skipRemoteRegistration,
			}

			// Register the user
			output, err := registerService.RegisterUser(ctx, input)
			if err != nil {
				fmt.Printf("Error during registration: %v\n", err)
				return
			}

			// Display success message
			if skipRemoteRegistration {
				fmt.Println("\n✅ Registration information saved locally.")
				fmt.Println("To complete registration with the remote server, run the command again without the --skip-remote flag.")
			} else {
				fmt.Println("\n✅ Registration successful!")
				if output.ServerResponse != "" {
					fmt.Println(output.ServerResponse)
				} else {
					fmt.Println("Please check your email for verification instructions.")
				}
				fmt.Println("\nIMPORTANT: Please ensure you have saved your password securely.")
				fmt.Println("You will need it to log in to your account.")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (required)")
	cmd.Flags().StringVarP(&firstName, "firstname", "f", "", "First name for the user (required)")
	cmd.Flags().StringVarP(&lastName, "lastname", "l", "", "Last name for the user (required)")
	cmd.Flags().StringVarP(&timezone, "timezone", "t", "UTC", "Timezone for the user (e.g., \"America/New_York\")")
	cmd.Flags().StringVarP(&country, "country", "c", "Canada", "Country for the user")
	cmd.Flags().StringVarP(&phone, "phone", "n", "", "Phone number for the user")
	cmd.Flags().StringVar(&betaAccessCode, "beta-code", "", "Beta access code (if required)")
	cmd.Flags().BoolVar(&agreeTerms, "agree-terms", false, "Agree to Terms of Service")
	cmd.Flags().BoolVar(&agreePromotions, "agree-promotions", false, "Agree to receive promotions")
	cmd.Flags().BoolVar(&agreeTracking, "agree-tracking", false, "Agree to tracking across third-party apps and services")
	cmd.Flags().IntVarP(&module, "module", "m", 0, "Module the user is registering for")
	cmd.Flags().BoolVar(&skipRemoteRegistration, "skip-remote", false, "Skip remote registration and only save locally")

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("firstname")
	cmd.MarkFlagRequired("lastname")

	return cmd
}
