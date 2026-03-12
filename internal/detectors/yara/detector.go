package yara

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
	backends []yara
}

func New(rulesDir string) (*YaraDetector, error) {
	yarac, err := newYarac()
	if err != nil {
		return nil, err
	}

	yarax, err := newYarax()
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(rulesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yar" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		name := filepath.Base(path)
		if err := yarax.compile(name, content); err != nil {
			if err := yarac.compile(name, content); err != nil {
				return fmt.Errorf("compile %s: %w", name, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", rulesDir, err)
	}

	for _, b := range []yara{yarax, yarac} {
		if err := b.build(); err != nil {
			return nil, err
		}
	}

	return &YaraDetector{
		backends: []yara{yarax, yarac},
	}, nil
}

func (d *YaraDetector) Name() string {
	return "yara"
}

func (d *YaraDetector) Scan(ctx context.Context, target string) ([]core.MatchResult, error) {
	f, err := os.Open(target)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", target, err)
	}
	defer f.Close()

	data, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		// fallback sang ScanFile nếu mmap thất bại
		return d.scanFile(ctx, target)
	}
	defer data.Unmap()

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
			matches, err := b.scanMem(ctx, target, data)
			ch <- result{matches, err}
		}(b)
	}

	wg.Wait()
	close(ch)

	var (
		all  []core.MatchResult
		errs []error
	)

	for r := range ch {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		all = append(all, r.matches...)
	}

	if len(errs) == len(d.backends) {
		return nil, fmt.Errorf("yara: all backends failed: %v", errs)
	}

	return all, nil
}

// fallback nếu mmap thất bại
func (d *YaraDetector) scanFile(ctx context.Context, target string) ([]core.MatchResult, error) {
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
			matches, err := b.scan(ctx, target)
			ch <- result{matches, err}
		}(b)
	}

	wg.Wait()
	close(ch)

	var (
		all  []core.MatchResult
		errs []error
	)

	for r := range ch {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		all = append(all, r.matches...)
	}

	if len(errs) == len(d.backends) {
		return nil, fmt.Errorf("yara: all backends failed: %v", errs)
	}

	return all, nil
}

func (d *YaraDetector) Close() {
	for _, b := range d.backends {
		b.close()
	}
}
