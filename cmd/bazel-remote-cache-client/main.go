package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type application struct {
	BazelRemoteCache *BazelRemoteCache
}

func (app *application) Cleanup() {
	if app.BazelRemoteCache != nil {
		app.BazelRemoteCache.Close()
	}
}

func main() {
	var app application
	defer app.Cleanup()

	var (
		remoteFlag       string
		instanceNameFlag string
		noColorFlag      bool
	)

	cmd := cobra.Command{
		Use:   "bazel-remote-cache-client",
		Short: "CLI to show Bazel remote cache entries (CA and CAS)",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if noColorFlag {
				disableColor()
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			app.BazelRemoteCache, err = NewBazelRemoteCache(
				ctx, remoteFlag, instanceNameFlag,
			)

			return err
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := cmd.PersistentFlags()
	flags.SortFlags = false

	flags.StringVarP(
		&remoteFlag, "remote", "r", "",
		"Remote cache URL (<host>:<port>)",
	)
	cmd.MarkPersistentFlagRequired("remote")

	flags.StringVarP(
		&instanceNameFlag, "instance-name", "i", "",
		"Instance name of the remote cache",
	)

	flags.BoolVarP(
		&noColorFlag, "no-color", "", false,
		"Disable color output",
	)

	flags.BoolP("help", "h", false, "Show this help and exit")

	cmd.AddCommand(
		newACCmd(&app),
	)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
