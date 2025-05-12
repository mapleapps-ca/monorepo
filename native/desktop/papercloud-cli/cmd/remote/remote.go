// native/desktop/papercloud-cli/cmd/remote/remote.go
package remote

import (
	"github.com/spf13/cobra"
)

func RemoteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remote",
		Short: "Execute commands related to making remote API calls",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing...
		},
	}

	// Add Remote-related commands
	cmd.AddCommand(HealthCheckCmd())
	cmd.AddCommand(RegisterUserCmd())

	return cmd
}
