package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	remoteexecution "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"github.com/spf13/cobra"
)

func newCASGetCmd(app *application) *cobra.Command {
	var (
		digest         *remoteexecution.Digest
		outputFilePath string
		isExecutable   bool
	)

	cmd := cobra.Command{
		Use:   "get [flags] <digest> ...",
		Short: "Get output file from remote Bazel cache",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			digest, err = parseDigestFromString(args[0])
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

	flags := cmd.Flags()
	flags.StringVarP(
		&outputFilePath, "output", "o", "",
		"Output file to write the blob",
	)
	flags.BoolVarP(
		&isExecutable, "exec", "x", false,
		"The blob content is executable",
	)

	return &cmd
}

func parseDigestFromString(s string) (*remoteexecution.Digest, error) {
	pair := strings.Split(s, "/")
	if len(pair) != 2 {
		return nil, fmt.Errorf("expected digest in the form hash/size, got %s", s)
	}

	size, err := strconv.ParseInt(pair[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size in digest %s: %s", s, err)
	}

	return &remoteexecution.Digest{
		Hash:      pair[0],
		SizeBytes: size,
	}, nil
}
