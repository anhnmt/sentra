package runner

import (
	"context"
	"encoding/json"
	"fmt"

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

	results, err := r.detector.Scan(ctx, r.opts.Target)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))

	return nil
}

func (r *Runner) Close() {
	fmt.Println("Goodbye!")
}
