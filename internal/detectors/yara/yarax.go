package yara

import (
	"context"
	"fmt"
	"sync"

	yarax "github.com/VirusTotal/yara-x/go"

	"github.com/anhnmt/sentra/internal/core"
)

type yaraxDetector struct {
	compiler    *yarax.Compiler
	rules       *yarax.Rules
	scannerPool sync.Pool
}

func newYarax() (*yaraxDetector, error) {
	compiler, err := yarax.NewCompiler()
	if err != nil {
		return nil, fmt.Errorf("yarax: new compiler: %w", err)
	}
	return &yaraxDetector{compiler: compiler}, nil
}

func (d *yaraxDetector) compile(name string, content []byte) error {
	if err := d.compiler.AddSource(string(content)); err != nil {
		return fmt.Errorf("yarax: compile %s: %w", name, err)
	}
	return nil
}

func (d *yaraxDetector) build() error {
	rules := d.compiler.Build()
	d.rules = rules

	// khởi tạo pool sau khi có rules
	d.scannerPool = sync.Pool{
		New: func() any {
			scanner := yarax.NewScanner(d.rules)
			return scanner
		},
	}

	return nil
}

func (d *yaraxDetector) scan(ctx context.Context, target string) ([]core.MatchResult, error) {
	if d.rules == nil {
		return nil, nil
	}

	scanner := yarax.NewScanner(d.rules)
	results, err := scanner.ScanFile(target)
	if err != nil {
		return nil, fmt.Errorf("yarax: scan %s: %w", target, err)
	}

	var all []core.MatchResult
	for _, m := range results.MatchingRules() {
		meta := make(map[string]interface{})
		for _, kv := range m.Metadata() {
			meta[kv.Identifier()] = kv.Value()
		}

		all = append(all, core.MatchResult{
			DetectorName: "yarax",
			RuleName:     m.Identifier(),
			Target:       target,
			Metadata:     meta,
		})
	}

	return all, nil
}

func (d *yaraxDetector) scanMem(ctx context.Context, target string, data []byte) ([]core.MatchResult, error) {
	if d.rules == nil {
		return nil, nil
	}

	scanner, ok := d.scannerPool.Get().(*yarax.Scanner)
	if !ok || scanner == nil {
		return nil, fmt.Errorf("yarax: failed to get scanner from pool")
	}
	defer d.scannerPool.Put(scanner) // trả về pool sau khi dùng xong

	results, err := scanner.Scan(data)
	if err != nil {
		return nil, fmt.Errorf("yarax: scanmem %s: %w", target, err)
	}

	var all []core.MatchResult
	for _, m := range results.MatchingRules() {
		meta := make(map[string]interface{})
		for _, kv := range m.Metadata() {
			meta[kv.Identifier()] = kv.Value()
		}
		all = append(all, core.MatchResult{
			DetectorName: "yarax",
			RuleName:     m.Identifier(),
			Target:       target,
			Metadata:     meta,
		})
	}

	return all, nil
}

func (d *yaraxDetector) close() {
	if d.rules != nil {
		d.rules.Destroy()
		d.rules = nil
	}
	if d.compiler != nil {
		d.compiler.Destroy()
		d.compiler = nil
	}
}
