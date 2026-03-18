package yara

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charlievieth/fastwalk"
	"github.com/edsrzf/mmap-go"

	"github.com/anhnmt/sentra/internal/core"
)

type yaraBackend interface {
	compile(name string, content []byte) error
	build() error
	scan(ctx context.Context, target string) ([]core.MatchResult, error)
	scanMem(ctx context.Context, target string, data []byte) ([]core.MatchResult, error)
	close()
}

type backendResult struct {
	matches []core.MatchResult
	err     error
}

type YaraDetector struct {
	backends    []yaraBackend
	absRulesDir string
	wg          sync.WaitGroup
}

func New(rulesDir string) (*YaraDetector, error) {
	absRulesDir, err := filepath.Abs(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("resolve rules dir: %w", err)
	}

	yaraxB, err := newYarax()
	if err != nil {
		return nil, err
	}

	yaracB, err := newYarac()
	if err != nil {
		return nil, err
	}

	if err := loadRules(rulesDir, yaraxB, yaracB); err != nil {
		return nil, err
	}

	for _, b := range []yaraBackend{yaraxB, yaracB} {
		if err := b.build(); err != nil {
			return nil, err
		}
	}

	return &YaraDetector{
		backends:    []yaraBackend{yaraxB, yaracB},
		absRulesDir: absRulesDir,
	}, nil
}

func loadRules(rulesDir string, backends ...yaraBackend) error {
	return fastwalk.Walk(&fastwalk.Config{Follow: false}, rulesDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".yar" {
				return err
			}
			return compileRule(path, backends...)
		},
	)
}

func compileRule(path string, backends ...yaraBackend) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	data, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return fmt.Errorf("mmap %s: %w", path, err)
	}
	defer data.Unmap()

	name := filepath.Base(path)
	var lastErr error
	for _, b := range backends {
		if err := b.compile(name, data); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func (d *YaraDetector) Name() string { return "yara" }

func (d *YaraDetector) Close() {
	d.wg.Wait()
	for _, b := range d.backends {
		b.close()
	}
}

func (d *YaraDetector) Scan(_ context.Context, target string, data []byte) ([]core.MatchResult, error) {
	d.wg.Add(1)
	defer d.wg.Done()
	return d.runBackends(target, data)
}

func (d *YaraDetector) runBackends(target string, data []byte) ([]core.MatchResult, error) {
	ch := make(chan backendResult, len(d.backends))
	var wg sync.WaitGroup

	for _, b := range d.backends {
		wg.Add(1)
		go func(b yaraBackend) {
			defer wg.Done()
			var m []core.MatchResult
			var err error
			if data != nil {
				m, err = b.scanMem(context.Background(), target, data)
			} else {
				m, err = b.scan(context.Background(), target)
			}
			ch <- backendResult{m, err}
		}(b)
	}

	wg.Wait()
	close(ch)

	return collectBackendResults(ch, len(d.backends))
}

func collectBackendResults(ch <-chan backendResult, total int) ([]core.MatchResult, error) {
	seen := make(map[string]struct{}, total)
	all := make([]core.MatchResult, 0, total)
	errCount := 0

	for r := range ch {
		if r.err != nil {
			errCount++
			continue
		}
		for _, m := range r.matches {
			key := m.RuleName + "|" + m.DetectorName
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, m)
		}
	}

	if errCount == total {
		return nil, fmt.Errorf("yara: all backends failed")
	}
	return all, nil
}

func (d *YaraDetector) IsRulesPath(target string) bool {
	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	return abs == d.absRulesDir ||
		strings.HasPrefix(abs, d.absRulesDir+string(os.PathSeparator))
}
