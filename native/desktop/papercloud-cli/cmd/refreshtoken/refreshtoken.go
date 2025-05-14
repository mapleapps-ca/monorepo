// monorepo/native/desktop/papercloud-cli/cmd/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RefreshTokenCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "refreshtoken",
		Short: "Force refresh current authenticated user's token",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Refreshing now...")
		},
	}
	return cmd
}
