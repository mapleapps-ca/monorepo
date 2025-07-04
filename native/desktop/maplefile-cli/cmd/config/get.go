// monorepo/native/desktop/maplefile-cli/cmd/config/get.go
package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
)

func getConfigCmd(configService config.ConfigService) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get current cloud provider address",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			address, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				fmt.Printf("Error getting cloud provider address: %v\n", err)
				return
			}
			fmt.Printf("Cloud Provider Address: %s\n", address)
		},
	}
}
