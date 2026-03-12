package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
)

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
}

func New(opts *Options) (*Runner, error) {
	detector, err := yara.New(opts.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("init yara: %w", err)
	}

	runner := &Runner{
		opts:     opts,
		detector: detector,
	}

	return runner, nil
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opts.Target == "" {
		return fmt.Errorf("--target is required")
	}

	defer r.detector.Close()

	var results []core.MatchResult

	err := filepath.WalkDir(r.opts.Target, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip signatures directory
		if d.IsDir() && path == r.opts.RulesDir {
			return filepath.SkipDir
		}

		// only scan regular file, skip symlink/dir
		if d.Type() != 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		matches, err := r.detector.Scan(ctx, path)
		if err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}

		results = append(results, matches...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", r.opts.Target, err)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))

	return nil
}

func (r *Runner) Close() {
	r.detector.Close()
	fmt.Println("Goodbye!")
}
