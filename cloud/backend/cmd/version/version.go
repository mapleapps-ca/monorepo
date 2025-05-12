// github.com/mapleapps-ca/monorepo/cloud/backend/cmd/version/version.go
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Major = "1"
const Minor = "0"
const Fix = "0"
const ReleaseType = "production"

// Configured via -ldflags during build
var GitCommit string

func VersionCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Describes version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fmt.Sprintf("Version: %s.%s.%s-%s", Major, Minor, Fix, ReleaseType))
		},
	}
	return cmd
}
