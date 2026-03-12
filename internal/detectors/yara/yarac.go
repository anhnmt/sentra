package yara

import (
	"context"
	"fmt"
	"sync"

	yarac "github.com/hillu/go-yara/v4"

	"github.com/anhnmt/sentra/internal/core"
)

type yaracDetector struct {
	compiler    *yarac.Compiler
	rules       *yarac.Rules
	scannerPool sync.Pool
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

	// khởi tạo pool sau khi có rules
	d.scannerPool = sync.Pool{
		New: func() any {
			scanner, err := yarac.NewScanner(d.rules)
			if err != nil {
				return nil
			}
			return scanner
		},
	}

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

func (d *yaracDetector) scanMem(ctx context.Context, target string, data []byte) ([]core.MatchResult, error) {
	if d.rules == nil {
		return nil, nil
	}

	scanner, ok := d.scannerPool.Get().(*yarac.Scanner)
	if !ok || scanner == nil {
		return nil, fmt.Errorf("yarac: failed to get scanner from pool")
	}
	defer d.scannerPool.Put(scanner)

	cb := &yaracCallback{target: target}
	scanner.SetCallback(cb)

	if err := scanner.ScanMem(data); err != nil {
		return nil, fmt.Errorf("yarac: scanmem %s: %w", target, err)
	}

	return cb.results, nil
}

func (d *yaracDetector) close() {
	if d.rules != nil {
		d.rules.Destroy()
		d.rules = nil
	}
	if d.compiler != nil {
		d.compiler.Destroy()
		d.compiler = nil
	}
}

type yaracCallback struct {
	target  string
	results []core.MatchResult
}

func (c *yaracCallback) RuleMatching(_ *yarac.ScanContext, r *yarac.Rule) (bool, error) {
	meta := make(map[string]string)
	for _, kv := range r.Metas() {
		meta[kv.Identifier] = fmt.Sprintf("%v", kv.Value)
	}

	c.results = append(c.results, core.MatchResult{
		DetectorName: "yarac",
		RuleName:     r.Identifier(),
		Target:       c.target,
		Metadata:     meta,
	})

	return true, nil // true = tiếp tục scan
}
