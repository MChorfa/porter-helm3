package main

import (
	"github.com/MChorfa/porter-helm3/pkg/helm3"
	"github.com/spf13/cobra"
)

func buildBuildCommand(m *helm3.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Generate Dockerfile lines for the bundle invocation image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Build(cmd.Context())
		},
	}
	return cmd
}
