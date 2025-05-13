// cmd/remote/config.go
package remote

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
	"github.com/spf13/cobra"
)

func ConfigCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Manage remote configuration",
		Long:  `Get or set the remote server configuration, such as the cloud provider address.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(getConfigCmd(configUseCase))
	cmd.AddCommand(setConfigCmd(configUseCase))

	return cmd
}

func getConfigCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get current cloud provider address",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			address, err := configUseCase.GetCloudProviderAddress(ctx)
			if err != nil {
				fmt.Printf("Error getting cloud provider address: %v\n", err)
				return
			}
			fmt.Printf("Cloud Provider Address: %s\n", address)
		},
	}
}

func setConfigCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "set [address]",
		Short: "Set cloud provider address",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			address := args[0]
			if err := configUseCase.SetCloudProviderAddress(ctx, address); err != nil {
				fmt.Printf("Error setting cloud provider address: %v\n", err)
				return
			}
			fmt.Printf("Cloud Provider Address set to: %s\n", address)
		},
	}
}
