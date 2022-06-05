package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/leboncoin/bazel-remote-cache-client/pkg/bzlremotecache"
)

func newCASGetCmd(app *application) *cobra.Command {
	var (
		digest         *bzlremotecache.Digest
		outputFilePath string
		isExecutable   bool
	)

	cmd := cobra.Command{
		Use:   "get [flags] <digest> ...",
		Short: "Get output file from remote Bazel cache",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			digest, err = bzlremotecache.ParseDigestFromString(args[0])
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := app.BazelRemoteCache.GetBlob(cmd.Context(), digest)
			if err != nil {
				return err
			}

			var output io.Writer
			if outputFilePath == "" {
				output = os.Stdout
			} else {
				var perm os.FileMode
				if isExecutable {
					perm = 0755
				} else {
					perm = 0644
				}

				f, err := os.OpenFile(
					outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm,
				)

				if err != nil {
					return fmt.Errorf("failed to open the output file: %v", err)
				}

				defer func() {
					if err := f.Close(); err != nil {
						_, _ = fmt.Fprintf(
							os.Stderr, "Warning: Can't close the output file: %v\n", err,
						)
					}
				}()

				output = f
			}

			outputBuf := bufio.NewWriter(output)

			if _, err := outputBuf.Write(content); err != nil {
				return fmt.Errorf("can't write the output file: %v", err)
			}

			if err := outputBuf.Flush(); err != nil {
				return fmt.Errorf("can't flush the output file: %v", err)
			}

			return nil
		},
	}

	fl := cmd.Flags()
	fl.StringVarP(
		&outputFilePath, "output", "o", "",
		"Output file to write the blob",
	)
	fl.BoolVarP(
		&isExecutable, "exec", "x", false,
		"The blob content is executable",
	)

	return app.newRemoteCacheCommand(&cmd)
}
