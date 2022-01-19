package main

import (
	"fmt"

	"github.com/fatih/color"

	"github.mpi-internal.com/jean-baptiste-bronisz/bazel-remote-cache-client/internal/bzlremotecache"
)

var (
	acDigestColor = color.New(color.FgYellow, color.Bold)
	errorColor    = color.New(color.FgRed, color.Bold)
	cyanColor     = color.New(color.FgCyan)
	boldColor     = color.New(color.Bold)
	faintColor    = color.New(color.Faint)
	redColor      = color.New(color.FgRed)
)

func disableColor() {
	color.NoColor = true
}

func printActionResult(prefix string, ar *bzlremotecache.ActionResult) {
	if len(ar.OutputFiles) > 0 {
		fmt.Print(prefix + boldColor.Sprint("OutputFiles") + ":\n")
		for _, outputFile := range ar.OutputFiles {
			printOutputFile(prefix+"  ", &outputFile)
		}
	}

	if ar.StdoutDigest != nil {
		fmt.Printf(prefix+"Stdout: %s\n", getColoredDigest(ar.StdoutDigest))
	}

	if ar.StderrDigest != nil {
		fmt.Printf(prefix+"Stderr: %s\n", getColoredDigest(ar.StderrDigest))
	}
}

func printOutputFile(prefix string, of *bzlremotecache.OutputFile) {
	var isExecutableMarker string
	if of.IsExecutable {
		isExecutableMarker = redColor.Sprint("x ")
	}

	fmt.Printf(prefix+"- %s%s\n", isExecutableMarker, cyanColor.Sprint(of.Path))
	fmt.Printf(prefix+"  |- %s\n", getColoredDigest(&of.Digest))
}

func getColoredDigest(d *bzlremotecache.Digest) string {
	return faintColor.Sprintf("%s/%d", d.Hash, d.Size)
}
