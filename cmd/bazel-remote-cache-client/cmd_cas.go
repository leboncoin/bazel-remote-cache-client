package main

import (
	"github.com/spf13/cobra"
)

func newCASCmd(app *application) *cobra.Command {
	cmd := cobra.Command{
		Use:     "content-addressable-store [flags]",
		Short:   "Manage CAS entries",
		Aliases: []string{"cas"},
	}

	cmd.AddCommand(
		newCASGetCmd(app),
	)

	return &cmd
}
