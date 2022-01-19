package bzlremotecache

// ActionResult is the action cache entry.
type ActionResult struct {
	OutputFiles  []OutputFile
	StdoutDigest *Digest
	StderrDigest *Digest
}

// OutputFile represents a output file of a aciont cache.
type OutputFile struct {
	Path         string
	Digest       Digest
	IsExecutable bool
}
