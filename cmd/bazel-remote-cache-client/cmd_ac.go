package main

import (
	"github.com/spf13/cobra"
)

func newACCmd(app *application) *cobra.Command {
	cmd := cobra.Command{
		Use:     "action-cache [flags]",
		Short:   "Manage AC entries",
		Aliases: []string{"ac"},
	}

	cmd.AddCommand(
		newACGetCmd(app),
	)

	return &cmd
}
