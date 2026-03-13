package runner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
	"github.com/rs/zerolog/log"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
	"github.com/anhnmt/sentra/internal/progress"
	"github.com/anhnmt/sentra/internal/worker"
)

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
	pool     *worker.Pool
}

func New(opts *Options) (*Runner, error) {
	detector, err := yara.New(opts.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("init yara: %w", err)
	}

	pool, err := worker.New(&worker.Options{Size: opts.Workers})
	if err != nil {
		return nil, fmt.Errorf("init worker pool: %w", err)
	}

	return &Runner{opts: opts, detector: detector, pool: pool}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opts.Target == "" {
		return fmt.Errorf("--target is required")
	}

	log.Info().
		Str("target", r.opts.Target).
		Str("rules_dir", r.opts.RulesDir).
		Int("workers", r.opts.Workers).
		Msg("scan starting")

	bar := progress.New(progress.Options{
		Workers: r.opts.Workers,
	})

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
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				bar.Increment(fileCount.Add(1))

				wg.Add(1)
				if err := r.pool.Submit(func() {
					defer wg.Done()
					matches, err := r.detector.Scan(ctx, path)
					ch <- result{matches, err}
				}); err != nil {
					wg.Done()
					return fmt.Errorf("submit %s: %w", path, err)
				}
				return nil
			},
		)

		wg.Wait()
		close(ch)
	}()

	<-done
	bar.Done(fileCount.Load())

	log.Info().
		Int64("files", fileCount.Load()).
		Int64("matches", matchCount.Load()).
		Int("errors", len(errs)).
		Dur("duration", time.Since(start)).
		Msg("scan complete")

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return fmt.Errorf("walk %s: %w", r.opts.Target, walkErr)
	}
	if len(errs) > 0 {
		return fmt.Errorf("scan errors: %v", errs)
	}
	return nil
}

func (r *Runner) Close() {
	r.pool.Close()
	r.detector.Close()
	log.Info().Msg("Goodbye!")
}
