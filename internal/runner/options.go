package runner

import (
	"runtime"

	"github.com/spf13/pflag"
)

type Options struct {
	UpdateSignatures bool
	RulesDir         string
	Target           string
	Workers          int
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
	fs.StringVar(&opts.Target, "target", "", "File or directory to scan")
	fs.IntVar(&opts.Workers, "workers", runtime.NumCPU()*2, "Number of worker goroutines")
	return fs
}
