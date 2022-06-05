package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/leboncoin/bazel-remote-cache-client/pkg/bzlremotecache"
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
		noColorFlag bool
	)

	cmd := cobra.Command{
		Use:   "bazel-remote-cache-client",
		Short: "CLI to show Bazel remote cache entries (CA and CAS)",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if noColorFlag {
				disableColor()
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       appVersion,
	}

	fl := cmd.PersistentFlags()
	fl.SortFlags = false

	fl.BoolVarP(
		&noColorFlag, "no-color", "", false,
		"Disable color output",
	)
	fl.BoolP("help", "h", false, "Show this help and exit")

	cmd.AddCommand(
		newACCmd(&app),
		newCASCmd(&app),
		newLogCmd(&app),
	)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (app *application) newRemoteCacheCommand(cmd *cobra.Command) *cobra.Command {
	var (
		remoteFlag       string
		instanceNameFlag string
	)

	oldPreRunE := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		var err error

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		if remoteFlag == "" {
			return errors.New("bazel remote cache address not given")
		}

		app.BazelRemoteCache, err = bzlremotecache.New(
			ctx, remoteFlag, instanceNameFlag,
		)

		if err != nil {
			return err
		}

		if oldPreRunE != nil {
			return oldPreRunE(cmd, args)
		}

		return nil
	}

	fl := cmd.Flags()
	fl.StringVarP(
		&remoteFlag, "remote", "r", os.Getenv("BAZEL_REMOTE_CACHE"),
		"Remote cache URL (<host>:<port>)",
	)
	fl.StringVarP(
		&instanceNameFlag, "instance-name", "i", "",
		"Instance name of the remote cache",
	)

	return cmd
}
