package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

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
