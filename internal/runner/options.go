package runner

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Name     string
	Host     string
	RulesDir string
	Target   string
}

func ParseOptions() *Options {
	opts := &Options{}

	NewGroups(pflag.CommandLine).
		Add("Database", NewDatabaseGroup(opts)).
		Add("Server", NewServerGroup(opts)).
		Add("Yara", NewYaraGroup(opts)).
		Parse()

	return opts
}

func NewDatabaseGroup(opts *Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("database", pflag.ExitOnError)
	fs.StringVar(&opts.Name, "name", "world", "Database host")
	return fs
}

func NewServerGroup(opts *Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("server", pflag.ExitOnError)
	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Server host")
	return fs
}

func NewYaraGroup(opts *Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("yara", pflag.ExitOnError)
	fs.StringVar(&opts.RulesDir, "rules-dir", "signatures/yara", "YARA rules directory")
	fs.StringVar(&opts.Target, "target", "", "File or directory to scan")
	return fs
}
