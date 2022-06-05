package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/leboncoin/bazel-remote-cache-client/pkg/bzlremotelogging"
)

func newLogCmd(_ *application) *cobra.Command {
	var (
		showMetadata bool
	)

	cmd := cobra.Command{
		Use:   "log [flags] <filepath>...",
		Short: "Print in a human-readable gRPC remote execution log files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var count int
			for _, logFilePath := range args {
				if len(args) > 1 {
					if count > 0 {
						fmt.Println()
					}

					fmt.Printf("%s\n------\n", logFilePath)
				}

				if err := printLogFile(logFilePath, showMetadata); err != nil {
					return err
				}

				count++
			}

			return nil
		},
		Example: `  To generate a log file:
	$ bazel build \
	    --remote_cache=grpc://localhost:9092 \
	    --experimental_remote_grpc_log=/tmp/grpc.log \
	    //...

  To parse this log file:
	$ bazel-remote-cache-client log /tmp/grpc.log`,
	}

	fl := cmd.Flags()
	fl.BoolVarP(
		&showMetadata, "show-metadata", "m", false,
		"Show metadata of all log entries",
	)

	return &cmd
}

func printLogFile(logFilePath string, showMetadata bool) error {
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("can't open file %q: %v", logFilePath, err)
	}

	defer func() {
		_ = logFile.Close()
	}()

	var count int
	return readStreamProtoLog(logFile, func(le *bzlremotelogging.LogEntry) {
		if count > 0 {
			fmt.Println()
		}

		printLogEntry(le, showMetadata)

		count++
	})
}

func readStreamProtoLog(r io.Reader, processLogFunc func(le *bzlremotelogging.LogEntry)) error {
	buf := make([]byte, 0, 4096)

	br := bufio.NewReader(r)
	for {
		size, err := binary.ReadUvarint(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if int(size) > cap(buf) {
			buf = make([]byte, 0, size)
		}
		buf = buf[:size]

		_, err = io.ReadFull(br, buf)
		if err != nil {
			return err
		}

		var le bzlremotelogging.LogEntry
		if err := proto.Unmarshal(buf, &le); err != nil {
			return err
		}

		processLogFunc(&le)
	}
}
