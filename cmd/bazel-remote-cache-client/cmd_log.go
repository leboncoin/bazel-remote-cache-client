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

	"github.mpi-internal.com/jean-baptiste-bronisz/bazel-remote-cache-client/pkg/bzlremotelogging"
)

func newLogCmd(_ *application) *cobra.Command {
	var (
		logFilePath string
	)

	cmd := cobra.Command{
		Use:   "log [flags]",
		Short: "Print in a human-readable a remote execution gRPC log file",
		RunE: func(cmd *cobra.Command, args []string) error {
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

				printLogEntry(le)

				count++
			})
		},
	}

	fl := cmd.Flags()
	fl.StringVar(
		&logFilePath, "file", "",
		"Path of the log file to parse",
	)

	return &cmd
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
