package main

import (
	"fmt"
	"strings"

	remoteexecution "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"github.com/bazelbuild/remote-apis/build/bazel/semver"
	"github.com/fatih/color"
	"google.golang.org/grpc/codes"

	"github.mpi-internal.com/jean-baptiste-bronisz/bazel-remote-cache-client/pkg/bzlremotelogging"
)

var (
	acDigestColor = color.New(color.FgYellow, color.Bold)
	errorColor    = color.New(color.FgRed, color.Bold)
	okColor       = color.New(color.FgGreen)
	cyanColor     = color.New(color.FgCyan)
	magentaColor  = color.New(color.FgMagenta, color.Bold)
	yellowColor   = color.New(color.FgYellow)
	boldColor     = color.New(color.Bold)
	faintColor    = color.New(color.Faint)
	redColor      = color.New(color.FgRed)

	reqPrefix  = color.New(color.FgGreen).Sprint("-->")
	respPrefix = color.New(color.FgRed).Sprint("<--")
)

func disableColor() {
	color.NoColor = true
}

func printActionResult(prefix string, ar *remoteexecution.ActionResult) {
	if len(ar.OutputFiles) > 0 {
		fmt.Printf(prefix+"%s:\n", cf("OutputFiles"))
		for _, outputFile := range ar.OutputFiles {
			printOutputFile(prefix+"  ", outputFile)
		}
	}

	if len(ar.OutputFileSymlinks) > 0 {
		fmt.Printf(prefix+"%s:\n", cf("OutputFileSymlinks"))
		for _, outputFileSymlink := range ar.OutputFileSymlinks {
			printOutputSymlink(prefix+"  ", outputFileSymlink)
		}
	}

	if len(ar.OutputSymlinks) > 0 {
		fmt.Printf(prefix+"%s:\n", cf("OutputSymlinks"))
		for _, outputSymlink := range ar.OutputSymlinks {
			printOutputSymlink(prefix+"  ", outputSymlink)
		}
	}

	if len(ar.OutputDirectories) > 0 {
		fmt.Printf(prefix+"%s:\n", cf("OutputDirectories"))
		for _, outputDirectory := range ar.OutputDirectories {
			printOutputDirectory(prefix+"  ", outputDirectory)
		}
	}

	if len(ar.OutputDirectorySymlinks) > 0 {
		fmt.Printf(prefix+"%s:\n", cf("OutputDirectorySymlinks"))
		for _, outputDirectorySymlink := range ar.OutputDirectorySymlinks {
			printOutputSymlink(prefix+"  ", outputDirectorySymlink)
		}
	}

	if ar.ExitCode != 0 {
		fmt.Printf(prefix+"%s: %d\n", cf("ExitCode"), ar.ExitCode)
	}

	if ar.StdoutDigest != nil && ar.StdoutDigest.SizeBytes > 0 {
		fmt.Printf(prefix+"%s: %s\n", cf("Stdout"), getColoredDigest(ar.StdoutDigest))
	}

	if ar.StderrDigest != nil && ar.StderrDigest.SizeBytes > 0 {
		fmt.Printf(prefix+"%s: %s\n", cf("Stderr"), getColoredDigest(ar.StderrDigest))
	}
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
	fmt.Printf(
		prefix+"- %s -> %s\n",
		cyanColor.Sprint(ofs.Path),
		cyanColor.Sprint(ofs.Target),
	)
}

func printOutputDirectory(prefix string, od *remoteexecution.OutputDirectory) {
	fmt.Printf(prefix+"- %s\n", cyanColor.Sprint(od.Path))
	fmt.Printf(prefix+"  |- %s\n", getColoredDigest(od.TreeDigest))
}

func printLogEntry(le *bzlremotelogging.LogEntry) {
	startTime := le.StartTime.AsTime().Local()
	endTime := le.EndTime.AsTime().Local()

	fmt.Printf(
		"[%s] %s - %s (%s)\n",
		startTime.Format("02 Jan 2006 15:04:05.000"),
		getColoredGRPCMethod(le.MethodName),
		getColoredGRPCCode(le.Status.Code),
		endTime.Sub(startTime),
	)

	prefix := "    "

	if le.Status.Message != "" {
		fmt.Printf(prefix+"%s: %s\n", cf("Message"), faintColor.Sprint(le.Status.Message))
	}

	fmt.Printf(prefix+"%s:\n", cf("Metadata"))
	fmt.Printf(prefix+"|- %s:\t\t\t%s (%s) / %s\n", cf("Tool"),
		le.Metadata.ToolDetails.ToolName,
		le.Metadata.ToolDetails.ToolVersion,
		le.Metadata.ToolInvocationId,
	)
	if le.Metadata.ActionId != "" {
		fmt.Printf(prefix+"|- %s:\t\t%s\n", cf("ActionId"), le.Metadata.ActionId)
	}
	if le.Metadata.CorrelatedInvocationsId != "" {
		fmt.Printf(prefix+"|- %s:\t%s\n", cf("CorrelatedInvocationsId"), le.Metadata.CorrelatedInvocationsId)
	}
	if le.Metadata.ActionMnemonic != "" {
		fmt.Printf(prefix+"|- %s:\t\t%s\n", cf("ActionMnemonic"), le.Metadata.ActionMnemonic)
	}
	if le.Metadata.TargetId != "" {
		fmt.Printf(prefix+"|- %s:\t\t%s\n", cf("TargetId"), magentaColor.Sprint(le.Metadata.TargetId))
	}
	if le.Metadata.ConfigurationId != "" {
		fmt.Printf(prefix+"|- %s:\t\t%s\n", cf("ConfigurationId"), le.Metadata.ConfigurationId)
	}

	if gcr := le.Details.GetGetCapabilities(); gcr != nil {
		printGetCapabilitiesDetails(prefix, gcr)
	} else if gar := le.Details.GetGetActionResult(); gar != nil {
		printGetActionResultDetails(prefix, gar)
	} else if r := le.Details.GetRead(); r != nil {
		printReadDetails(prefix, r)
	} else if fmb := le.Details.GetFindMissingBlobs(); fmb != nil {
		printFindMissingBlobsDetails(prefix, fmb)
	} else if w := le.Details.GetWrite(); w != nil {
		printWriteDetails(prefix, w)
	} else if uar := le.Details.GetUpdateActionResult(); uar != nil {
		printUpdateActionResultDetails(prefix, uar)
	}
}

func printGetCapabilitiesDetails(prefix string, gcr *bzlremotelogging.GetCapabilitiesDetails) {
	fmt.Println(prefix + reqPrefix)
	if gcr.Request.InstanceName != "" {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("InstanceName"), gcr.Request.InstanceName)
	}

	fmt.Println(prefix + respPrefix)
	if gcr.Response.CacheCapabilities != nil {
		fmt.Printf(prefix+"\t|- %s: %+v\n", cf("CacheCapabilities"), gcr.Response.CacheCapabilities)
	}
	if gcr.Response.ExecutionCapabilities != nil {
		fmt.Printf(prefix+"\t|- %s: %+v\n", cf("ExecutionCapabilities"), gcr.Response.ExecutionCapabilities)
	}
	if gcr.Response.DeprecatedApiVersion != nil {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("DeprecatedApiVersion"), sv(gcr.Response.DeprecatedApiVersion))
	}
	if gcr.Response.LowApiVersion != nil {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("LowApiVersion"), sv(gcr.Response.LowApiVersion))
	}
	if gcr.Response.HighApiVersion != nil {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("HighApiVersion"), sv(gcr.Response.HighApiVersion))
	}
}

func printGetActionResultDetails(prefix string, gar *bzlremotelogging.GetActionResultDetails) {
	fmt.Println(prefix + reqPrefix)
	if gar.Request.InstanceName != "" {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("InstanceName"), gar.Request.InstanceName)
	}
	if gar.Request.ActionDigest != nil {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("ActionDigest"), getColoredDigest(gar.Request.ActionDigest))
	}
	if gar.Request.InlineStdout {
		fmt.Printf(prefix+"\t|- %s: %t\n", cf("InlineStdout"), gar.Request.InlineStdout)
	}
	if gar.Request.InlineStderr {
		fmt.Printf(prefix+"\t|- %s: %t\n", cf("InlineStderr"), gar.Request.InlineStderr)
	}
	if len(gar.Request.InlineOutputFiles) > 0 {
		fmt.Printf(prefix+"\t|- %s:\n", cf("InlineOutputFiles"))
		for _, iof := range gar.Request.InlineOutputFiles {
			fmt.Printf(prefix+"\t   - %s:\n", iof)
		}
	}

	fmt.Println(prefix + respPrefix)
	if gar.Response != nil {
		printActionResult(prefix+"\t|- ", gar.Response)
	}
}

func printUpdateActionResultDetails(prefix string, uar *bzlremotelogging.UpdateActionResultDetails) {
	fmt.Println(prefix + reqPrefix)
	if uar.Request.InstanceName != "" {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("InstanceName"), uar.Request.InstanceName)
	}
	if uar.Request.ActionDigest != nil && uar.Request.ActionDigest.SizeBytes > 0 {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("ActionDigest"), getColoredDigest(uar.Request.ActionDigest))
	}
	if uar.Request.ActionResult != nil {
		fmt.Printf(prefix+"\t|- %s:\n", cf("ActionResult"))
		printActionResult(prefix+"\t\t|- ", uar.Request.ActionResult)
	}
	if uar.Request.ResultsCachePolicy != nil {
		fmt.Printf(prefix+"\t|- %s:\n", cf("ResultsCachePolicy"))
		fmt.Printf(prefix+"\t\t|- %s: %d\n", cf("Priority"), uar.Request.ResultsCachePolicy.Priority)
	}

	fmt.Println(prefix + respPrefix)
	if uar.Response != nil {
		printActionResult(prefix+"\t|- ", uar.Response)
	}
}

func printReadDetails(prefix string, r *bzlremotelogging.ReadDetails) {
	fmt.Println(prefix + reqPrefix)
	if r.Request.ResourceName != "" {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("ResourceName"), faintColor.Sprint(r.Request.ResourceName))
	}
	if r.Request.ReadOffset > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("ReadOffset"), r.Request.ReadOffset)
	}
	if r.Request.ReadLimit > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("ReadLimit"), r.Request.ReadLimit)
	}

	fmt.Println(prefix + respPrefix)
	if r.NumReads > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("NumReads"), r.NumReads)
	}
	if r.BytesRead > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("BytesRead"), r.BytesRead)
	}
}

func printWriteDetails(prefix string, w *bzlremotelogging.WriteDetails) {
	fmt.Println(prefix + reqPrefix)
	if len(w.ResourceNames) > 0 {
		fmt.Printf(prefix+"\t|- %s\n", cf("ResourceNames"))
		for _, resourceName := range w.ResourceNames {
			fmt.Printf(prefix+"\t   - %s\n", faintColor.Sprint(resourceName))
		}
	}
	if len(w.Offsets) > 0 {
		fmt.Printf(prefix+"\t|- %s\n", cf("Offsets"))
		for _, offset := range w.Offsets {
			fmt.Printf(prefix+"\t   - %d\n", offset)
		}
	}
	if len(w.FinishWrites) > 0 {
		fmt.Printf(prefix+"\t|- %s\n", cf("FinishWrites"))
		for _, finishWrite := range w.FinishWrites {
			fmt.Printf(prefix+"\t   - %d\n", finishWrite)
		}
	}
	if w.NumWrites > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("NumWrites"), w.NumWrites)
	}
	if w.BytesSent > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("BytesSent"), w.BytesSent)
	}

	fmt.Println(prefix + respPrefix)
	if w.Response.CommittedSize > 0 {
		fmt.Printf(prefix+"\t|- %s: %d\n", cf("CommittedSize"), w.Response.CommittedSize)
	}
}

func printFindMissingBlobsDetails(prefix string, fmb *bzlremotelogging.FindMissingBlobsDetails) {
	fmt.Println(prefix + reqPrefix)
	if fmb.Request.InstanceName != "" {
		fmt.Printf(prefix+"\t|- %s: %s\n", cf("InstanceName"), fmb.Request.InstanceName)
	}
	if len(fmb.Request.BlobDigests) > 0 {
		fmt.Printf(prefix+"\t|- %s\n", cf("BlobDigests"))
		for _, bd := range fmb.Request.BlobDigests {
			fmt.Printf(prefix+"\t   - %s\n", getColoredDigest(bd))
		}
	}

	fmt.Println(prefix + respPrefix)
	if len(fmb.Response.MissingBlobDigests) > 0 {
		fmt.Printf(prefix+"\t|- %s\n", cf("MissingBlobDigests"))
		for _, mbd := range fmb.Response.MissingBlobDigests {
			fmt.Printf(prefix+"\t   - %s\n", getColoredDigest(mbd))
		}
	}
}

func cf(name string) string {
	return boldColor.Sprint(name)
}

func sv(v *semver.SemVer) string {
	if v.Patch == 0 {
		return fmt.Sprintf("%d.%d%s", v.Major, v.Minor, v.Prerelease)
	}

	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Prerelease)
}

func getColoredDigest(d *remoteexecution.Digest) string {
	return faintColor.Sprintf("%s/%d", d.Hash, d.SizeBytes)
}

func getColoredGRPCCode(rawCode int32) string {
	code := codes.Code(rawCode)

	var c *color.Color
	switch code {
	case codes.OK:
		c = okColor
	case codes.NotFound:
		c = faintColor
	default:
		c = redColor
	}

	return c.Sprint(code.String())
}

func getColoredGRPCMethod(fullMethod string) string {
	pkgMethod, funcName, _ := strings.Cut(fullMethod, "/")
	return cyanColor.Sprint(pkgMethod) + "::" + yellowColor.Sprint(funcName)
}
