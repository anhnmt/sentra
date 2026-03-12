// internal/detectors/yara/yarac.go
package yara

import (
	"context"
	"fmt"

	yarac "github.com/hillu/go-yara/v4"

	"github.com/anhnmt/sentra/internal/core"
)

type yaracDetector struct {
	compiler *yarac.Compiler
	rules    *yarac.Rules
}

func newYarac() (*yaracDetector, error) {
	compiler, err := yarac.NewCompiler()
	if err != nil {
		return nil, fmt.Errorf("yarac: new compiler: %w", err)
	}
	return &yaracDetector{compiler: compiler}, nil
}

func (d *yaracDetector) compile(name string, content []byte) error {
	if err := d.compiler.AddString(string(content), name); err != nil {
		return fmt.Errorf("yarac: compile %s: %w", name, err)
	}
	return nil
}

func (d *yaracDetector) build() error {
	rules, err := d.compiler.GetRules()
	if err != nil {
		return fmt.Errorf("yarac: build: %w", err)
	}
	d.rules = rules
	return nil
}

func (d *yaracDetector) scan(ctx context.Context, target string) ([]core.MatchResult, error) {
	if d.rules == nil {
		return nil, nil
	}

	var matches yarac.MatchRules
	if err := d.rules.ScanFile(target, 0, 0, &matches); err != nil {
		return nil, fmt.Errorf("yarac: scan %s: %w", target, err)
	}

	all := make([]core.MatchResult, 0, len(matches))
	for _, m := range matches {
		meta := make(map[string]string)
		for _, kv := range m.Metas {
			meta[kv.Identifier] = fmt.Sprintf("%v", kv.Value)
		}

		all = append(all, core.MatchResult{
			DetectorName: "yarac",
			RuleName:     m.Rule,
			Target:       target,
			Metadata:     meta,
		})
	}

	return all, nil
}
