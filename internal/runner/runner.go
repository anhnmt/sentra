package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charlievieth/fastwalk"
	"github.com/edsrzf/mmap-go"
	"github.com/rs/zerolog/log"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
	"github.com/anhnmt/sentra/internal/logger"
	"github.com/anhnmt/sentra/internal/progress"
	"github.com/anhnmt/sentra/internal/worker"
)

const (
	mmapThreshold = 512 * 1024 // 512KB
	batchSize     = 16
)

// defaultSkipDirs — built-in, luôn được skip
var defaultSkipDirs = map[string]struct{}{
	".git":         {},
	".svn":         {},
	"node_modules": {},
	"vendor":       {},
	".cache":       {},
	".devenv":      {},
}

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
	ioPool   *worker.Pool
	scanPool *worker.Pool
	skipDirs map[string]struct{}
}

func New(opts *Options) (*Runner, error) {
	detector, err := yara.New(opts.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("init yara: %w", err)
	}

	ioPool, err := worker.New(&worker.Options{Size: opts.Workers})
	if err != nil {
		return nil, fmt.Errorf("init io pool: %w", err)
	}

	scanPool, err := worker.NewScanPool()
	if err != nil {
		return nil, fmt.Errorf("init scan pool: %w", err)
	}

	skipDirs := make(map[string]struct{}, len(defaultSkipDirs)+len(opts.SkipDirs))
	for k := range defaultSkipDirs {
		skipDirs[k] = struct{}{}
	}
	for _, d := range opts.SkipDirs {
		skipDirs[d] = struct{}{}
	}

	return &Runner{
		opts:     opts,
		detector: detector,
		ioPool:   ioPool,
		scanPool: scanPool,
		skipDirs: skipDirs,
	}, nil
}

type scanItem struct {
	path    string
	data    []byte
	cleanup func()
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opts.Target == "" {
		return fmt.Errorf("--target is required")
	}

	if abs, err := filepath.Abs(r.opts.Target); err == nil {
		r.opts.Target = abs
	}

	log.Info().
		Str("target", r.opts.Target).
		Str("rules_dir", r.opts.RulesDir).
		Int("workers", r.opts.Workers).
		Int("max_depth", r.opts.MaxDepth).
		Strs("skip_dirs", r.opts.SkipDirs).
		Msg("scan starting")

	bar := progress.New(progress.Options{
		Workers: r.opts.Workers,
	})
	logger.InitWithWriter(bar.Writer)

	type result struct {
		matches []core.MatchResult
		err     error
	}

	ch := make(chan result, r.opts.Workers)

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		errs       []error
		errorCount int64
		done       = make(chan struct{})
	)

	start := time.Now()

	// consumer
	go func() {
		defer close(done)
		for res := range ch {
			if res.err != nil {
				if !errors.Is(res.err, os.ErrPermission) {
					errorCount++
					mu.Lock()
					errs = append(errs, res.err)
					mu.Unlock()
				}
				continue
			}
			for _, match := range res.matches {
				bar.IncrementMatch()
				log.Warn().
					Str("detector", match.DetectorName).
					Str("rule", match.RuleName).
					Str("file", match.Target).
					Fields(match.Metadata).
					Msg("match detected")
			}
		}
	}()

	submitBatch := func(batch []scanItem) {
		if len(batch) == 0 {
			return
		}
		wg.Add(1)
		items := batch
		if err := r.scanPool.Submit(func() {
			defer wg.Done()
			for _, item := range items {
				matches, err := r.detector.Scan(context.Background(), item.path, item.data)
				item.cleanup()
				ch <- result{matches, err}
			}
		}); err != nil {
			wg.Done()
			for _, item := range items {
				item.cleanup()
				ch <- result{nil, fmt.Errorf("submit scan %s: %w", item.path, err)}
			}
		}
	}

	baseDepth := strings.Count(r.opts.Target, string(os.PathSeparator))

	var walkErr error
	go func() {
		var (
			batchMu sync.Mutex
			batch   []scanItem
		)

		flushBatch := func() {
			batchMu.Lock()
			cur := batch
			batch = nil
			batchMu.Unlock()
			submitBatch(cur)
		}

		walkErr = fastwalk.Walk(&fastwalk.Config{Follow: false}, r.opts.Target,
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					if errors.Is(err, os.ErrPermission) {
						return nil
					}
					return err
				}

				if d.IsDir() {
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

				if d.Type() != 0 {
					return nil
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}

				info, err := d.Info()
				if err != nil {
					return nil
				}
				if !yara.IsEligibleInfo(info, r.opts.MinFileSize, r.opts.MaxFileSize) {
					bar.IncrementSkip()
					return nil
				}
				if r.detector.IsRulesPath(path) {
					return nil
				}

				bar.IncrementFile()

				isMmap := info.Size() >= mmapThreshold

				wg.Add(1)
				if err := r.ioPool.Submit(func() {
					data, cleanup, err := readFile(path)
					if err != nil {
						wg.Done()
						ch <- result{nil, err}
						return
					}
					if data == nil {
						cleanup()
						wg.Done()
						return
					}
					if yara.HasSkipMagic(data) {
						cleanup()
						wg.Done()
						bar.IncrementSkip()
						return
					}

					if isMmap {
						if err := r.scanPool.Submit(func() {
							defer wg.Done()
							defer cleanup()
							matches, err := r.detector.Scan(context.Background(), path, data)
							ch <- result{matches, err}
						}); err != nil {
							cleanup()
							wg.Done()
							ch <- result{nil, fmt.Errorf("submit scan %s: %w", path, err)}
						}
						return
					}

					batchMu.Lock()
					batch = append(batch, scanItem{path: path, data: data, cleanup: cleanup})
					full := len(batch) >= batchSize
					cur := batch
					if full {
						batch = nil
					}
					batchMu.Unlock()

					wg.Done()
					if full {
						submitBatch(cur)
					}
				}); err != nil {
					wg.Done()
					return fmt.Errorf("submit io %s: %w", path, err)
				}
				return nil
			},
		)

		flushBatch()
		wg.Wait()
		close(ch)
	}()

	<-done
	bar.Done()
	logger.InitWithWriter(bar.Writer)

	canceled := errors.Is(ctx.Err(), context.Canceled)

	if canceled {
		log.Warn().
			Int64("scanned", bar.Files()).
			Int64("skipped", bar.Skipped()).
			Int64("matches", bar.Matches()).
			Str("duration", time.Since(start).Round(time.Second).String()).
			Msg("scan canceled")
	} else {
		log.Info().
			Int64("scanned", bar.Files()).
			Int64("skipped", bar.Skipped()).
			Int64("matches", bar.Matches()).
			Int64("errors", errorCount).
			Str("duration", time.Since(start).Round(time.Second).String()).
			Msg("scan complete")
	}

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return fmt.Errorf("walk %s: %w", r.opts.Target, walkErr)
	}
	if len(errs) > 0 && !canceled {
		return fmt.Errorf("scan errors: %v", errs)
	}
	return nil
}

func (r *Runner) Close() {
	r.scanPool.Close()
	r.ioPool.Close()
	r.detector.Close()
	log.Info().Msg("Goodbye!")
}

func readFile(path string) ([]byte, func(), error) {
	noop := func() {}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, noop, nil
		}
		return nil, noop, fmt.Errorf("stat %s: %w", path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, noop, nil
		}
		return nil, noop, fmt.Errorf("open %s: %w", path, err)
	}

	if info.Size() < mmapThreshold {
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, noop, fmt.Errorf("read %s: %w", path, err)
		}
		return data, noop, nil
	}

	mapped, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		f.Close()
		f2, err2 := os.Open(path)
		if err2 != nil {
			return nil, noop, fmt.Errorf("open %s: %w", path, err2)
		}
		defer f2.Close()
		data, err2 := io.ReadAll(f2)
		if err2 != nil {
			return nil, noop, fmt.Errorf("read %s: %w", path, err2)
		}
		return data, noop, nil
	}

	return mapped, func() { mapped.Unmap(); f.Close() }, nil
}
