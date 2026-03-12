package runner

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type Groups struct {
	cmd     *pflag.FlagSet
	entries []GroupConfig // dùng GroupConfig luôn, không cần struct riêng
}

type GroupConfig struct {
	Display string
	FlagSet *pflag.FlagSet
}

func NewGroups(cmd *pflag.FlagSet) *Groups {
	return &Groups{cmd: cmd}
}

func (g *Groups) Add(display string, fs *pflag.FlagSet) *Groups {
	g.entries = append(g.entries, GroupConfig{display, fs})
	g.cmd.AddFlagSet(fs)
	return g
}

func (g *Groups) Parse() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		for _, e := range g.entries {
			fmt.Fprintf(os.Stderr, "%s:\n", e.Display)
			e.FlagSet.PrintDefaults() // in flag của từng group riêng
			fmt.Fprintln(os.Stderr)
		}
	}

	pflag.Parse()
}
