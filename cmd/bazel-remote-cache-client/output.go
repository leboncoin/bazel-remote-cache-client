package main

import (
	"fmt"

	remoteexecution "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"github.com/fatih/color"
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

func printActionResult(prefix string, ar *remoteexecution.ActionResult) {
	if len(ar.OutputFiles) > 0 {
		fmt.Print(prefix + boldColor.Sprint("OutputFiles") + ":\n")
		for _, outputFile := range ar.OutputFiles {
			printOutputFile(prefix+"  ", outputFile)
		}
	}

	if len(ar.OutputFileSymlinks) > 0 {
		fmt.Printf(prefix + boldColor.Sprint("OutputFileSymlinks") + ":\n")
		for _, ofs := range ar.OutputFileSymlinks {
			printOutputSymlink(prefix+"  ", ofs)
		}
	}

	if len(ar.OutputSymlinks) > 0 {
		fmt.Printf(prefix + boldColor.Sprint("OutputSymlinks") + ":\n")
		for _, os := range ar.OutputSymlinks {
			printOutputSymlink(prefix+"  ", os)
		}
	}

	if len(ar.OutputDirectories) > 0 {
		fmt.Printf(prefix + boldColor.Sprint("OutputDirectories") + ":\n")
		for _, od := range ar.OutputDirectories {
			printOutputDirectory(prefix+"  ", od)
		}
	}

	if len(ar.OutputDirectorySymlinks) > 0 {
		fmt.Printf(prefix + boldColor.Sprint("OutputDirectorySymlinks") + ":\n")
		for _, ods := range ar.OutputDirectorySymlinks {
			printOutputSymlink(prefix+"  ", ods)
		}
	}

	printDigestIfNeeded(prefix, boldColor.Sprint("Stdout"), ar.StdoutDigest)
	printDigestIfNeeded(prefix, boldColor.Sprint("Stderr"), ar.StderrDigest)
}

func printOutputFile(prefix string, of *remoteexecution.OutputFile) {
	var isExecutableMarker string
	if of.IsExecutable {
		isExecutableMarker = redColor.Sprint("x ")
	}

	fmt.Printf(prefix+"- %s%s\n", isExecutableMarker, cyanColor.Sprint(of.Path))
	fmt.Printf(prefix+"  |- %s\n", getColoredDigest(of.Digest))
}

func printOutputSymlink(prefix string, ofs *remoteexecution.OutputSymlink) {
	fmt.Printf(prefix+"- %s\n", cyanColor.Sprint(ofs.Path))
	fmt.Printf(prefix+"  -> %s\n", ofs.Target)
}

func printOutputDirectory(prefix string, od *remoteexecution.OutputDirectory) {
	fmt.Printf(prefix+"- %s\n", cyanColor.Sprint(od.Path))
	fmt.Printf(prefix+"  |- %s\n", getColoredDigest(od.TreeDigest))
}

func printDigestIfNeeded(prefix string, name string, d *remoteexecution.Digest) {
	if d == nil || d.SizeBytes == 0 {
		return
	}

	printDigest(prefix, name, d)
}

func printDigest(prefix string, name string, d *remoteexecution.Digest) {
	fmt.Printf(prefix+"%s: %s\n", name, getColoredDigest(d))
}

func getColoredDigest(d *remoteexecution.Digest) string {
	return faintColor.Sprintf("%s/%d", d.Hash, d.SizeBytes)
}
