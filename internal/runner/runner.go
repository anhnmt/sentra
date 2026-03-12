package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/charlievieth/fastwalk"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
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

	runner := &Runner{
		opts:     opts,
		detector: detector,
		pool:     pool,
	}

	return runner, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opts.Target == "" {
		return fmt.Errorf("--target is required")
	}

	type result struct {
		matches []core.MatchResult
		err     error
	}

	ch := make(chan result, r.opts.Workers)
	var wg sync.WaitGroup

	// stream writer — JSON encode từng result ngay khi có
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	var (
		errs []error
		mu   sync.Mutex
		done = make(chan struct{})
	)

	// collector + streamer goroutine
	go func() {
		defer close(done)
		for r := range ch {
			if r.err != nil {
				mu.Lock()
				errs = append(errs, r.err)
				mu.Unlock()
				continue
			}

			for _, match := range r.matches {
				mu.Lock()
				_ = encoder.Encode(match)
				mu.Unlock()
			}
		}
	}()

	var walkErr error
	go func() {
		walkErr = fastwalk.Walk(&fastwalk.Config{
			Follow: false,
		}, r.opts.Target, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
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
		})

		wg.Wait()
		close(ch)
	}()

	<-done

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

	fmt.Println("Goodbye!")
}
