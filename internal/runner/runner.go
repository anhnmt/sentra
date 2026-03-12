package runner

import (
	"context"
	"fmt"
)

type Runner struct {
	opts *Options
}

func New(opts *Options) (*Runner, error) {
	runner := &Runner{
		opts: opts,
	}

	return runner, nil
}

func (r *Runner) Run(ctx context.Context) error {
	fmt.Printf("Runner starting...\n")
	fmt.Printf("Hello %s", r.opts.Name)

	return nil
}

func (r *Runner) Close() {
	fmt.Println("Goodbye!")
}
