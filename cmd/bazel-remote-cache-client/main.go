package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.mpi-internal.com/jean-baptiste-bronisz/bazel-remote-cache-client/internal/bzlremotecache"
)

var (
	appVersion = "unknown"
)

type application struct {
	BazelRemoteCache *bzlremotecache.BazelRemoteCache
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

			if remoteFlag == "" {
				return errors.New("bazel remote cache address not given")
			}

			app.BazelRemoteCache, err = bzlremotecache.New(
				ctx, remoteFlag, instanceNameFlag,
			)

			return err
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       appVersion,
	}

	flags := cmd.PersistentFlags()
	flags.SortFlags = false

	flags.StringVarP(
		&remoteFlag, "remote", "r", os.Getenv("BAZEL_REMOTE_CACHE"),
		"Remote cache URL (<host>:<port>)",
	)

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
		newCASCmd(&app),
	)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
