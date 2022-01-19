package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newACGetCmd(app *application) *cobra.Command {
	return &cobra.Command{
		Use:   "get [flags] <digest> ...",
		Short: "Get action result metadata from Bazel remote cache",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				hasError       bool
				hasCacheResult bool
			)

			for _, digest := range args {
				result, err := app.BazelRemoteCache.GetCacheResult(cmd.Context(), digest)

				if hasCacheResult {
					fmt.Println()
				}

				if err != nil {
					fmt.Printf(
						"%s: %s\n",
						acDigestColor.Sprint(digest),
						errorColor.Sprint(app.BazelRemoteCache.ErrorMsg(err)),
					)
					hasError = true
				} else {
					fmt.Printf("%s:\n", acDigestColor.Sprint(digest))
					printActionResult("  ", result)
				}

				hasCacheResult = true
			}

			if hasError {
				return errors.New("all action result hasn't been retrieved")
			}

			return nil
		},
	}
}
