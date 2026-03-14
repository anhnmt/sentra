package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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

const mmapThreshold = 512 * 1024 // 512KB

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
	ioPool   *worker.Pool // walk + open file + mmap
	scanPool *worker.Pool // CGo/Rust scan, giới hạn NumCPU
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

	return &Runner{
		opts:     opts,
		detector: detector,
		ioPool:   ioPool,
		scanPool: scanPool,
	}, nil
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
		fileCount  atomic.Int64
		matchCount atomic.Int64
		done       = make(chan struct{})
	)

	start := time.Now()

	// consumer
	go func() {
		defer close(done)
		for res := range ch {
			if res.err != nil {
				if !errors.Is(res.err, os.ErrPermission) {
					mu.Lock()
					errs = append(errs, res.err)
					mu.Unlock()
				}
				continue
			}
			for _, match := range res.matches {
				matchCount.Add(1)
				log.Warn().
					Str("detector", match.DetectorName).
					Str("rule", match.RuleName).
					Str("file", match.Target).
					Fields(match.Metadata).
					Msg("match detected")
			}
		}
	}()

	// producer
	var walkErr error
	go func() {
		walkErr = fastwalk.Walk(&fastwalk.Config{Follow: false}, r.opts.Target,
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					if errors.Is(err, os.ErrPermission) {
						return nil
					}
					return err
				}
				if d.IsDir() && path == r.opts.RulesDir {
					return fastwalk.SkipDir
				}
				if d.Type() != 0 {
					return nil
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}

				bar.Increment(fileCount.Add(1))

				wg.Add(1)
				if err := r.ioPool.Submit(func() {
					// ioPool: chỉ lo đọc file
					data, cleanup, err := readFile(path)
					if err != nil {
						wg.Done()
						ch <- result{nil, err}
						return
					}
					if data == nil {
						// file không đọc được (permission, v.v.) — skip
						cleanup()
						wg.Done()
						return
					}

					// scanPool: CGo/Rust scan
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
				}); err != nil {
					wg.Done()
					return fmt.Errorf("submit io %s: %w", path, err)
				}
				return nil
			},
		)

		wg.Wait()
		close(ch)
	}()

	<-done
	bar.Done(fileCount.Load())
	logger.InitWithWriter(bar.Writer)

	canceled := errors.Is(ctx.Err(), context.Canceled)

	if canceled {
		log.Warn().
			Int64("files", fileCount.Load()).
			Int64("matches", matchCount.Load()).
			Str("duration", time.Since(start).Round(time.Second).String()).
			Msg("scan canceled")
	} else {
		log.Info().
			Int64("files", fileCount.Load()).
			Int64("matches", matchCount.Load()).
			Int("errors", len(errs)).
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

// readFile đọc file với threshold:
// - nhỏ hơn mmapThreshold → io.ReadAll (ít overhead hơn)
// - lớn hơn → mmap (tránh copy vào heap)
// cleanup phải được gọi sau khi scan xong
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
		// fallback về ReadAll nếu mmap fail
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

	cleanup := func() {
		mapped.Unmap()
		f.Close()
	}
	return mapped, cleanup, nil
}
