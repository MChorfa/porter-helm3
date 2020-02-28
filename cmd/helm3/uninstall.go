package main

import (
	"github.com/MChorfa/porter-helm3/pkg/helm3"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(m *helm3.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Uninstall()
		},
	}
	return cmd
}
