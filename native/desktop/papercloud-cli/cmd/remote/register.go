// native/desktop/papercloud-cli/cmd/remote/register.go
package remote

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RegisterUserCmd() *cobra.Command {
	var email, password, firstName, lastName, timezone, country, phone, betaAccessCode string
	var agreeTerms, agreePromotions, agreeTracking bool
	var module int

	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Register user account",
		Long: `Register a new user account in the system.

This command requires you to provide an email, password, first name, and last name.
You can optionally provide timezone, country, phone number, a beta access code,
specify agreement to terms, promotions, and tracking, and specify the registration module.

Examples:
		# Register with only required fields
		register --email user@example.com --password mysecret --firstname John --lastname Doe

		# Register with all fields using short flags (note: only some have short flags)
		register -e test@domain.com -p pass123 -f Jane -l Smith -t "America/Toronto" -c "USA" -n "555-1234" --beta-code ABCDE --agree-terms --module 2

		# Register using a mix of short and long flags, enabling all agreements and specifying module
		register --email another@user.net -p anotherpass -f Bob -l Williams --timezone "Europe/London" --agree-terms --agree-promotions --agree-tracking --module 1`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Registering...")
			//TODO: IMPLEMENT THIS STUB.
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

	// Mark required flags
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("firstname")
	cmd.MarkFlagRequired("lastname")

	return cmd
}
