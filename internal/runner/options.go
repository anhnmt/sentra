package runner

import (
	"runtime"

	"github.com/projectdiscovery/goflags"
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
	SkipDirs         goflags.StringSlice
	MaxDepth         int
	MinFileSize      int
	MaxFileSize      int
	DBPath           string
	OutputPath       string
	ScanID           string
}

func ParseOptions() (*Options, error) {
	opts := &Options{
		DBPath: "sentra.db",
	}

	fs := goflags.NewFlagSet()
	fs.SetDescription("Sentra - threat intelligence scanner")

	fs.CreateGroup("scan", "Scan Options",
		fs.StringVar(&opts.Target, "target", "./", "file or directory to scan"),
		fs.StringVar(&opts.RulesDir, "rules-dir", "signatures/yara", "YARA rules directory"),
		fs.IntVar(&opts.Workers, "workers", runtime.NumCPU()*2, "number of I/O worker goroutines"),
		fs.StringSliceVar(&opts.SkipDirs, "skip-dir", goflags.StringSlice{".git", "node_modules", "vendor"}, "directory names to skip (repeatable)", goflags.CommaSeparatedStringSliceOptions),
		fs.IntVar(&opts.MaxDepth, "max-depth", 0, "maximum directory depth (0 = unlimited)"),
		fs.IntVar(&opts.MinFileSize, "min-file-size", defaultMinFileSize, "minimum file size to scan in bytes"),
		fs.IntVar(&opts.MaxFileSize, "max-file-size", defaultMaxFileSize, "maximum file size to scan in bytes"),
	)

	fs.CreateGroup("update", "Util",
		fs.BoolVar(&opts.UpdateSignatures, "update-signatures", false, "update latest signatures"),
	)

	fs.CreateGroup("store", "Database",
		fs.StringVar(&opts.DBPath, "db", opts.DBPath, "path to bbolt database file"),
	)

	fs.CreateGroup("report", "Report",
		fs.StringVar(&opts.OutputPath, "output", ".", "output HTML report path"),
		fs.StringVar(&opts.ScanID, "scan-id", "", "scan ID to generate report (latest if empty)"),
	)

	return opts, fs.Parse()
}
