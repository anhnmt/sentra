package worker

import (
	"fmt"

	"github.com/panjf2000/ants/v2"
)

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

func (p *Pool) Submit(task func()) error {
	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}
	return p.pool.Submit(task)
}

func (p *Pool) Close() {
	p.pool.Release()
}
