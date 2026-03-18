package runner

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

	"github.com/anhnmt/sentra/internal/detectors/yara"
	"github.com/anhnmt/sentra/internal/progress"
)

// batcher collects scanItems and flushes when full.
type batcher struct {
	mu    sync.Mutex
	items []scanItem
}

func (b *batcher) add(item scanItem) ([]scanItem, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = append(b.items, item)
	if len(b.items) < batchSize {
		return nil, false
	}
	cur := b.items
	b.items = nil
	return cur, true
}

func (b *batcher) flush() []scanItem {
	b.mu.Lock()
	defer b.mu.Unlock()
	cur := b.items
	b.items = nil
	return cur
}

func (r *Runner) walkFiles(ctx context.Context, ch chan<- scanResult, wg *sync.WaitGroup, bar *progress.Bar) error {
	baseDepth := strings.Count(r.opts.Target, string(os.PathSeparator))
	bat := &batcher{}

	err := fastwalk.Walk(&fastwalk.Config{Follow: false}, r.opts.Target,
		func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if errors.Is(walkErr, os.ErrPermission) {
					return nil
				}
				return walkErr
			}
			if d.IsDir() {
				return r.skipDir(path, baseDepth)
			}
			if d.Type() != 0 || ctx.Err() != nil {
				return ctx.Err()
			}
			return r.enqueueFile(path, d, ch, wg, bar, bat)
		},
	)

	r.submitBatch(bat.flush(), ch, wg)
	return err
}

func (r *Runner) skipDir(path string, baseDepth int) error {
	if path == r.opts.RulesDir {
		return fastwalk.SkipDir
	}
	if _, skip := r.skipDirs[filepath.Base(path)]; skip {
		return fastwalk.SkipDir
	}
	if r.opts.MaxDepth > 0 {
		depth := strings.Count(path, string(os.PathSeparator)) - baseDepth
		if depth >= r.opts.MaxDepth {
			return fastwalk.SkipDir
		}
	}
	return nil
}

func (r *Runner) enqueueFile(path string, d fs.DirEntry, ch chan<- scanResult, wg *sync.WaitGroup, bar *progress.Bar, bat *batcher) error {
	info, err := d.Info()
	if err != nil {
		return nil
	}
	if !yara.IsEligibleInfo(info, int64(r.opts.MinFileSize), int64(r.opts.MaxFileSize)) {
		bar.IncrementSkip()
		return nil
	}
	if r.detector.IsRulesPath(path) {
		return nil
	}
	bar.IncrementFile()

	large := info.Size() >= mmapThreshold
	wg.Add(1)
	if err := r.ioPool.Submit(func() {
		r.readAndEnqueue(path, large, ch, wg, bar, bat)
	}); err != nil {
		wg.Done()
		return fmt.Errorf("submit io %s: %w", path, err)
	}
	return nil
}

func (r *Runner) readAndEnqueue(path string, large bool, ch chan<- scanResult, wg *sync.WaitGroup, bar *progress.Bar, bat *batcher) {
	data, cleanup, err := readFile(path)
	if err != nil {
		wg.Done()
		ch <- scanResult{nil, err}
		return
	}
	if data == nil || yara.HasSkipMagic(data) {
		cleanup()
		wg.Done()
		if data != nil {
			bar.IncrementSkip()
		}
		return
	}
	if large {
		r.submitMmapScan(path, data, cleanup, ch, wg)
		return
	}
	wg.Done()
	if cur, full := bat.add(scanItem{path: path, data: data, cleanup: cleanup}); full {
		r.submitBatch(cur, ch, wg)
	}
}

func (r *Runner) submitMmapScan(path string, data []byte, cleanup func(), ch chan<- scanResult, wg *sync.WaitGroup) {
	if err := r.scanPool.Submit(func() {
		defer wg.Done()
		defer cleanup()
		matches, err := r.detector.Scan(context.Background(), path, data)
		ch <- scanResult{matches, err}
	}); err != nil {
		cleanup()
		wg.Done()
		ch <- scanResult{nil, fmt.Errorf("submit scan %s: %w", path, err)}
	}
}

func (r *Runner) submitBatch(batch []scanItem, ch chan<- scanResult, wg *sync.WaitGroup) {
	if len(batch) == 0 {
		return
	}
	wg.Add(1)
	if err := r.scanPool.Submit(func() {
		defer wg.Done()
		for _, item := range batch {
			matches, err := r.detector.Scan(context.Background(), item.path, item.data)
			item.cleanup()
			ch <- scanResult{matches, err}
		}
	}); err != nil {
		wg.Done()
		for _, item := range batch {
			item.cleanup()
			ch <- scanResult{nil, fmt.Errorf("submit scan %s: %w", item.path, err)}
		}
	}
}
