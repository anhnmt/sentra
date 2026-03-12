package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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

	ch := make(chan result, r.opts.Workers) // buffered = số worker
	var wg sync.WaitGroup

	var (
		all  []core.MatchResult
		errs []error
		mu   sync.Mutex
		done = make(chan struct{})
	)

	// collector goroutine
	go func() {
		defer close(done)
		for r := range ch {
			mu.Lock()
			if r.err != nil {
				errs = append(errs, r.err)
			} else {
				all = append(all, r.matches...)
			}
			mu.Unlock()
		}
	}()

	// walk goroutine — không block main
	var walkErr error
	go func() {
		walkErr = filepath.WalkDir(r.opts.Target, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && path == r.opts.RulesDir {
				return filepath.SkipDir
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

		// walk xong → chờ worker → close channel
		wg.Wait()
		close(ch)
	}()

	// chờ collector xong
	<-done

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return fmt.Errorf("walk %s: %w", r.opts.Target, walkErr)
	}

	if len(errs) > 0 {
		return fmt.Errorf("scan errors: %v", errs)
	}

	out, _ := json.MarshalIndent(all, "", "  ")
	fmt.Println(string(out))

	return nil
}

func (r *Runner) Close() {
	r.detector.Close()
	r.pool.Close()

	fmt.Println("Goodbye!")
}
