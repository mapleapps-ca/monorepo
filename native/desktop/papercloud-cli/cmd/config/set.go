// monorepo/native/desktop/papercloud-cli/cmd/config/set.go
package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

func setConfigCmd(configService config.ConfigService) *cobra.Command {
	return &cobra.Command{
		Use:   "set [address]",
		Short: "Set cloud provider address",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			address := args[0]
			if err := configService.SetCloudProviderAddress(ctx, address); err != nil {
				fmt.Printf("Error setting cloud provider address: %v\n", err)
				return
			}
			fmt.Printf("Cloud Provider Address set to: %s\n", address)
		},
	}
}
