package main

import (
	"github.com/MChorfa/porter-helm3/pkg/helm3"
	"github.com/spf13/cobra"
)

func buildSchemaCommand(m *helm3.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the json schema for the mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.PrintSchema()
		},
	}
	return cmd
}
