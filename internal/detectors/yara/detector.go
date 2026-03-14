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

type yara interface {
	compile(name string, content []byte) error
	build() error
	scan(ctx context.Context, target string) ([]core.MatchResult, error)
	scanMem(ctx context.Context, target string, data []byte) ([]core.MatchResult, error)
	close()
}

type YaraDetector struct {
	backends    []yara
	absRulesDir string
	wg          sync.WaitGroup
}

func New(rulesDir string) (*YaraDetector, error) {
	absRulesDir, err := filepath.Abs(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("resolve rules dir: %w", err)
	}

	yarax, err := newYarax()
	if err != nil {
		return nil, err
	}

	yarac, err := newYarac()
	if err != nil {
		return nil, err
	}

	if err := loadRules(rulesDir, yarax, yarac); err != nil {
		return nil, err
	}

	for _, b := range []yara{yarax, yarac} {
		if err := b.build(); err != nil {
			return nil, err
		}
	}

	return &YaraDetector{
		backends:    []yara{yarax, yarac},
		absRulesDir: absRulesDir,
	}, nil
}

func loadRules(rulesDir string, yarax, yarac yara) error {
	return fastwalk.Walk(&fastwalk.Config{Follow: false}, rulesDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".yar" {
				return err
			}
			return compileRule(path, yarax, yarac)
		},
	)
}

func compileRule(path string, yarax, yarac yara) error {
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
	if err := yarax.compile(name, data); err == nil {
		return nil
	}
	return yarac.compile(name, data)
}

func (d *YaraDetector) Name() string { return "yara" }

func (d *YaraDetector) Close() {
	d.wg.Wait()
	for _, b := range d.backends {
		b.close()
	}
}

func (d *YaraDetector) Scan(ctx context.Context, target string) ([]core.MatchResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if d.isRulesPath(target) || !isEligible(target) {
		return nil, nil
	}

	d.wg.Add(1)
	defer d.wg.Done()

	f, err := os.Open(target)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", target, err)
	}
	defer f.Close()

	data, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return d.runBackends(target, nil)
	}
	defer data.Unmap()

	return d.runBackends(target, data)
}

func (d *YaraDetector) runBackends(target string, data mmap.MMap) ([]core.MatchResult, error) {
	type result struct {
		matches []core.MatchResult
		err     error
	}

	ch := make(chan result, len(d.backends))
	var wg sync.WaitGroup

	for _, b := range d.backends {
		wg.Add(1)
		go func(b yara) {
			defer wg.Done()
			var m []core.MatchResult
			var err error
			if data != nil {
				m, err = b.scanMem(context.Background(), target, data)
			} else {
				m, err = b.scan(context.Background(), target)
			}
			ch <- result{m, err}
		}(b)
	}

	wg.Wait()
	close(ch)

	var all []core.MatchResult
	var errCount int
	for r := range ch {
		if r.err != nil {
			errCount++
			continue
		}
		all = append(all, r.matches...)
	}

	if errCount == len(d.backends) {
		return nil, fmt.Errorf("yara: all backends failed")
	}
	return all, nil
}

func (d *YaraDetector) isRulesPath(target string) bool {
	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	return abs == d.absRulesDir ||
		strings.HasPrefix(abs, d.absRulesDir+string(os.PathSeparator))
}
