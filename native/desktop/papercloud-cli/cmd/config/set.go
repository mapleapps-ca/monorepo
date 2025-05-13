package config

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
	"github.com/spf13/cobra"
)

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
