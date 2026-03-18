package worker

import (
	"fmt"
	"runtime"

	"github.com/panjf2000/ants/v2"
)

type Options struct {
	Size int
}

type Pool struct {
	pool *ants.Pool
}

func New(opts *Options) (*Pool, error) {
	pool, err := ants.NewPool(opts.Size,
		ants.WithPreAlloc(true),
		ants.WithNonblocking(false),
	)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}
	return &Pool{pool: pool}, nil
}

func NewScanPool() (*Pool, error) {
	pool, err := ants.NewPool(runtime.NumCPU(),
		ants.WithPreAlloc(true),
		ants.WithNonblocking(false),
		// Rust panic không propagate lên Go — cần recover để tránh crash
		ants.WithPanicHandler(func(v interface{}) {}),
	)
	if err != nil {
		return nil, fmt.Errorf("new scan pool: %w", err)
	}
	return &Pool{pool: pool}, nil
}

func (p *Pool) Submit(task func()) error {
	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}
	return p.pool.Submit(task)
}

func (p *Pool) Close() {
	p.pool.Release()
}
