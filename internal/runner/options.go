package runner

import (
	"runtime"

	"github.com/spf13/pflag"
)

const (
	defaultMinFileSize = 4 << 10  // 4KB
	defaultMaxFileSize = 30 << 20 // 30MB
)

type Options struct {
	UpdateSignatures bool
	Target           string
	RulesDir         string
	Workers          int
	ScanWorkers      int
	SkipDirs         []string
	MaxDepth         int
	MinFileSize      int64
	MaxFileSize      int64
}

func ParseOptions() *Options {
	opts := &Options{}

	NewGroups(pflag.CommandLine).
		Add("Scan options", NewScanOptionsGroup(opts)).
		Add("Util", NewUtilGroup(opts)).
		Parse()

	return opts
}

func NewUtilGroup(opts *Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("util", pflag.ExitOnError)
	fs.BoolVar(&opts.UpdateSignatures, "update-signatures", false, "Update latest signatures")
	return fs
}

func NewScanOptionsGroup(opts *Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("yara", pflag.ExitOnError)
	fs.StringVar(&opts.RulesDir, "rules-dir", "signatures/yara", "YARA rules directory")
	fs.StringVar(&opts.Target, "target", "./", "File or directory to scan")
	fs.IntVar(&opts.Workers, "workers", runtime.NumCPU()*2, "Number of I/O worker goroutines")
	fs.StringArrayVar(&opts.SkipDirs, "skip-dir", nil, "Directory names to skip (can be repeated: --skip-dir .git --skip-dir build)")
	fs.IntVar(&opts.MaxDepth, "max-depth", 0, "Maximum directory depth to walk (0 = unlimited)")
	fs.Int64Var(&opts.MinFileSize, "min-file-size", defaultMinFileSize, "Minimum file size to scan in bytes")
	fs.Int64Var(&opts.MaxFileSize, "max-file-size", defaultMaxFileSize, "Maximum file size to scan in bytes")
	return fs
}
