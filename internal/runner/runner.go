package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
	"github.com/anhnmt/sentra/internal/logger"
	"github.com/anhnmt/sentra/internal/progress"
	"github.com/anhnmt/sentra/internal/worker"
)

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
	ioPool   *worker.Pool
	scanPool *worker.Pool
	skipDirs map[string]struct{}
}

type scanResult struct {
	matches []core.MatchResult
	err     error
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

	skipDirs := make(map[string]struct{}, len(opts.SkipDirs))
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

	bar := progress.New(progress.Options{Workers: r.opts.Workers})
	logger.InitWithWriter(bar.Writer)
	start := time.Now()

	ch := make(chan scanResult, r.opts.Workers)
	var wg sync.WaitGroup

	errs, errCount, done := r.startConsumer(ch, bar)

	var walkErr error
	go func() {
		walkErr = r.walkFiles(ctx, ch, &wg, bar)
		wg.Wait()
		close(ch)
	}()

	<-done
	bar.Done()
	logger.InitWithWriter(bar.Writer)
	r.logScanSummary(ctx, bar, start, *errCount)

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return fmt.Errorf("walk %s: %w", r.opts.Target, walkErr)
	}
	if len(*errs) > 0 && !errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("scan errors: %v", *errs)
	}
	return nil
}

func (r *Runner) startConsumer(ch <-chan scanResult, bar *progress.Bar) (*[]error, *int64, <-chan struct{}) {
	var (
		mu         sync.Mutex
		errs       []error
		errorCount int64
		done       = make(chan struct{})
	)

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

	return &errs, &errorCount, done
}

func (r *Runner) logScanSummary(ctx context.Context, bar *progress.Bar, start time.Time, errorCount int64) {
	dur := time.Since(start).Round(time.Second).String()
	canceled := errors.Is(ctx.Err(), context.Canceled)

	if canceled {
		log.Warn().
			Int64("scanned", bar.Files()).
			Int64("skipped", bar.Skipped()).
			Int64("matches", bar.Matches()).
			Str("duration", dur).
			Msg("scan canceled")
		return
	}

	log.Info().
		Int64("scanned", bar.Files()).
		Int64("skipped", bar.Skipped()).
		Int64("matches", bar.Matches()).
		Int64("errors", errorCount).
		Str("duration", dur).
		Msg("scan complete")
}

func (r *Runner) Close() {
	r.scanPool.Close()
	r.ioPool.Close()
	r.detector.Close()
	log.Info().Msg("Goodbye!")
}
