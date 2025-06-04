// native/desktop/maplefile-cli/cmd/me/me.go
package me

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_me "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/me"
)

// MeCmd creates the main me command with subcommands
func MeCmd(
	getMeService svc_me.GetMeService,
	updateMeService svc_me.UpdateMeService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "me",
		Short: "Manage your user profile",
		Long:  `View and manage your user profile information.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add me subcommands
	cmd.AddCommand(getCmd(getMeService, logger))
	cmd.AddCommand(updateCmd(updateMeService, logger))

	return cmd
}
