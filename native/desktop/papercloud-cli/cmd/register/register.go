package register

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
)

func RegisterCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Execute commands related to registering an account on papercloud",
		Run: func(cmd *cobra.Command, args []string) {
			//TODO :IMPL.
		},
	}

	return cmd
}
